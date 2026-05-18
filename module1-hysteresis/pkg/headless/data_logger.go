package headless

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fecim-lattice-tools/shared/logging"
)

const (
	hysteresisDataLogBuffer = 8192
	mvPerCm                 = 1e8
	ucPerCm2                = 1e2
	// Downsample CSV logging by simulation time to avoid huge files and UI stalls.
	// Default: 250ms => ~4 samples/sec at most.
	// Override with FECIM_HYSTERESIS_LOG_INTERVAL_MS (float, milliseconds).
	hysteresisDataLogMinSimInterval = 2.5e-1

	// During ISPP write operations, use a much finer recording interval so that the
	// CSV log captures the actual E-field ramp and polarization trajectory instead of
	// showing apparent "teleportation" jumps between coarsely-sampled records.
	hysteresisDataLogISPPInterval = 1e-2 // 10ms => ~100 samples/sec during ISPP
)

type HysteresisDataLogger struct {
	path   string
	file   *os.File
	rows   chan HysteresisSnapshot
	wg     sync.WaitGroup
	closed uint32
	step   uint64

	minSimInterval  float64
	lastSimTimeBits uint64

	dropped     uint64
	lastDropLog int64
}

type HysteresisSnapshot struct {
	Step          uint64
	Timestamp     string
	SimTime       float64
	Dt            float64
	Waveform      string
	AutoMode      bool
	Material      string
	TemperatureK  float64
	EcMVcm        float64
	PsUcCm2       float64
	PrUcCm2       float64
	NumLevels     int
	LevelIndex    int
	Level         int
	StateBand     string
	EField        float64
	EFieldMVcm    float64
	Polarization  float64
	PolarizationU float64
	NormalizedP   float64

	WrdPhase       int
	WrdPhaseName   string
	WrdPhaseTimer  float64
	WrdTargetLevel int
	WrdReadLevel   int
	WrdRetryCount  int
	WrdCycleEnergy float64
	WrdTotalWrites int
	WrdSuccess     int
	WrdWriteE      float64
	WrdPrepE       float64
	WrdSettleE     float64
	WrdStartLevel  int

	ControllerState          string
	ControllerPhaseTimer     float64
	ControllerTargetLevel    int
	ControllerCurrentField   float64
	ControllerCurrentFieldMV float64
	ControllerPulseCount     int
	ControllerTotalPulses    int
	ControllerRetryCount     int
	ControllerOvershootCount int
	ControllerOvershootTotal int
	ControllerLastVerify     int
	ControllerLastError      int
	ControllerVMin           float64
	ControllerVMax           float64
	ControllerVMinEc         float64
	ControllerVMaxEc         float64
	ControllerInitialLevel   int
	ControllerFromSaturation bool
	ControllerResetDirection int
}

func NewHysteresisDataLogger(materialName string) (*HysteresisDataLogger, error) {
	if materialName == "" {
		materialName = "unknown"
	}
	logsDir := logging.LogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("create data log dir: %w", err)
	}

	// Include microseconds to avoid filename collisions when multiple headless
	// runs start within the same second (common in fast test suites).
	timestamp := time.Now().Format("2006-01-02_15-04-05.000000")
	safeMaterial := sanitizeMaterialName(materialName)
	filename := fmt.Sprintf("hysteresis-%s-%s.csv", safeMaterial, timestamp)
	path := filepath.Join(logsDir, filename)

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create data log file: %w", err)
	}

	minInterval := hysteresisDataLogMinSimInterval
	if v := strings.TrimSpace(os.Getenv("FECIM_HYSTERESIS_LOG_INTERVAL_MS")); v != "" {
		if ms, err := strconv.ParseFloat(v, 64); err == nil {
			if ms <= 0 {
				minInterval = 0
			} else {
				minInterval = ms / 1000.0
			}
		}
	}

	logger := &HysteresisDataLogger{
		path: path,
		file: file,
		rows: make(chan HysteresisSnapshot, hysteresisDataLogBuffer),
		// Throttle CSV logging to keep runtime smooth and files manageable.
		minSimInterval:  minInterval,
		lastSimTimeBits: math.Float64bits(-1),
	}
	logger.wg.Add(1)
	go logger.run()
	return logger, nil
}

func (l *HysteresisDataLogger) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

func (l *HysteresisDataLogger) shouldRecord(simTime float64) bool {
	return l.shouldRecordAt(simTime, l.minSimInterval)
}

// shouldRecordAt is like shouldRecord but accepts a custom minimum interval.
// Use a shorter interval for high-resolution phases (e.g., ISPP writes).
func (l *HysteresisDataLogger) shouldRecordAt(simTime float64, minInterval float64) bool {
	if l == nil || atomic.LoadUint32(&l.closed) == 1 {
		return false
	}
	if minInterval <= 0 {
		return true
	}
	lastBits := atomic.LoadUint64(&l.lastSimTimeBits)
	last := math.Float64frombits(lastBits)
	if last >= 0 && simTime >= last && (simTime-last) < minInterval {
		return false
	}
	atomic.StoreUint64(&l.lastSimTimeBits, math.Float64bits(simTime))
	return true
}

func (l *HysteresisDataLogger) Record(snapshot HysteresisSnapshot) {
	if l == nil || atomic.LoadUint32(&l.closed) == 1 {
		return
	}
	snapshot.Step = atomic.AddUint64(&l.step, 1)

	defer func() {
		_ = recover()
	}()
	if isCriticalHysteresisSnapshot(snapshot) {
		l.rows <- snapshot
		return
	}
	select {
	case l.rows <- snapshot:
		return
	default:
		atomic.AddUint64(&l.dropped, 1)
		now := time.Now().UnixNano()
		last := atomic.LoadInt64(&l.lastDropLog)
		if last != 0 && now-last < int64(time.Second) {
			return
		}
		if atomic.CompareAndSwapInt64(&l.lastDropLog, last, now) {
			dropped := atomic.LoadUint64(&l.dropped)
			logging.Printf("Hysteresis data log backlog: dropped %d samples", dropped)
		}
	}
}

func isCriticalHysteresisSnapshot(snapshot HysteresisSnapshot) bool {
	return snapshot.Waveform == "ISPP" &&
		snapshot.WrdTargetLevel > 0 &&
		snapshot.WrdPhaseTimer == 0
}

func (l *HysteresisDataLogger) Close() error {
	if l == nil {
		return nil
	}
	if !atomic.CompareAndSwapUint32(&l.closed, 0, 1) {
		return nil
	}
	close(l.rows)
	l.wg.Wait()
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("close data log file: %w", err)
	}
	return nil
}

func (l *HysteresisDataLogger) run() {
	defer l.wg.Done()

	writer := csv.NewWriter(l.file)
	if err := writer.Write(hysteresisDataHeader()); err != nil {
		logging.Printf("Hysteresis data log header error: %v", err)
		writer.Flush()
		return
	}
	writer.Flush()

	rowCount := 0
	for snapshot := range l.rows {
		if err := writer.Write(snapshot.toCSVRow()); err != nil {
			logging.Printf("Hysteresis data log write error: %v", err)
			continue
		}
		rowCount++
		if rowCount%256 == 0 {
			writer.Flush()
			if err := writer.Error(); err != nil {
				logging.Printf("Hysteresis data log flush error: %v", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		logging.Printf("Hysteresis data log final flush error: %v", err)
	}
}

func hysteresisDataHeader() []string {
	return []string{
		"step",
		"timestamp",
		"sim_time_s",
		"dt_s",
		"waveform",
		"auto_mode",
		"material",
		"temperature_k",
		"ec_mv_cm",
		"ps_uc_cm2",
		"pr_uc_cm2",
		"num_levels",
		"level_index",
		"level",
		"state_band",
		"e_field_v_m",
		"e_field_mv_cm",
		"polarization_c_m2",
		"polarization_uc_cm2",
		"normalized_p",
		"wrd_phase",
		"wrd_phase_name",
		"wrd_phase_timer_s",
		"wrd_target_level",
		"wrd_read_level",
		"wrd_retry_count",
		"wrd_cycle_energy_fj",
		"wrd_total_writes",
		"wrd_success_writes",
		"wrd_write_e_v_m",
		"wrd_prep_e_v_m",
		"wrd_settle_e_v_m",
		"wrd_start_level",
		"controller_state",
		"controller_phase_timer_s",
		"controller_target_level",
		"controller_current_field_v_m",
		"controller_current_field_mv_cm",
		"controller_pulse_count",
		"controller_total_pulses",
		"controller_retry_count",
		"controller_overshoot_count",
		"controller_overshoot_total",
		"controller_last_verify_level",
		"controller_last_error",
		"controller_vmin_v_m",
		"controller_vmax_v_m",
		"controller_vmin_ec",
		"controller_vmax_ec",
		"controller_initial_level",
		"controller_from_saturation",
		"controller_reset_direction",
	}
}

func (s HysteresisSnapshot) toCSVRow() []string {
	return []string{
		strconv.FormatUint(s.Step, 10),
		s.Timestamp,
		formatFloat(s.SimTime),
		formatFloat(s.Dt),
		s.Waveform,
		strconv.FormatBool(s.AutoMode),
		s.Material,
		formatFloat(s.TemperatureK),
		formatFloat(s.EcMVcm),
		formatFloat(s.PsUcCm2),
		formatFloat(s.PrUcCm2),
		strconv.Itoa(s.NumLevels),
		strconv.Itoa(s.LevelIndex),
		strconv.Itoa(s.Level),
		s.StateBand,
		formatFloat(s.EField),
		formatFloat(s.EFieldMVcm),
		formatFloat(s.Polarization),
		formatFloat(s.PolarizationU),
		formatFloat(s.NormalizedP),
		strconv.Itoa(s.WrdPhase),
		s.WrdPhaseName,
		formatFloat(s.WrdPhaseTimer),
		strconv.Itoa(s.WrdTargetLevel),
		strconv.Itoa(s.WrdReadLevel),
		strconv.Itoa(s.WrdRetryCount),
		formatFloat(s.WrdCycleEnergy),
		strconv.Itoa(s.WrdTotalWrites),
		strconv.Itoa(s.WrdSuccess),
		formatFloat(s.WrdWriteE),
		formatFloat(s.WrdPrepE),
		formatFloat(s.WrdSettleE),
		strconv.Itoa(s.WrdStartLevel),
		s.ControllerState,
		formatFloat(s.ControllerPhaseTimer),
		strconv.Itoa(s.ControllerTargetLevel),
		formatFloat(s.ControllerCurrentField),
		formatFloat(s.ControllerCurrentFieldMV),
		strconv.Itoa(s.ControllerPulseCount),
		strconv.Itoa(s.ControllerTotalPulses),
		strconv.Itoa(s.ControllerRetryCount),
		strconv.Itoa(s.ControllerOvershootCount),
		strconv.Itoa(s.ControllerOvershootTotal),
		strconv.Itoa(s.ControllerLastVerify),
		strconv.Itoa(s.ControllerLastError),
		formatFloat(s.ControllerVMin),
		formatFloat(s.ControllerVMax),
		formatFloat(s.ControllerVMinEc),
		formatFloat(s.ControllerVMaxEc),
		strconv.Itoa(s.ControllerInitialLevel),
		strconv.FormatBool(s.ControllerFromSaturation),
		strconv.Itoa(s.ControllerResetDirection),
	}
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func sanitizeMaterialName(name string) string {
	if name == "" {
		return "unknown"
	}
	safe := strings.ToLower(name)
	safe = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		case r == ' ' || r == '.' || r == '/':
			return '-'
		default:
			return -1
		}
	}, safe)
	if safe == "" {
		return "unknown"
	}
	return safe
}

func wrdPhaseName(phase int) string {
	switch phase {
	case 0:
		return "PREP"
	case 1:
		return "SETTLE"
	case 2:
		return "PROG_VERIFY"
	case 3:
		return "HOLD"
	case 4:
		return "READBACK"
	case 5:
		return "RESULT"
	case 6:
		return "RETRY"
	default:
		return "UNKNOWN"
	}
}

func stateBand(levelIndex int, numLevels int) string {
	if numLevels <= 0 {
		return "UNKNOWN"
	}
	lowThird := numLevels / 3
	highThird := numLevels * 2 / 3
	if levelIndex < lowThird {
		return "NEGATIVE"
	}
	if levelIndex >= highThird {
		return "POSITIVE"
	}
	return "INTERMEDIATE"
}
