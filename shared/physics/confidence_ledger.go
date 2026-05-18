package physics

import (
	"strings"

	"fecim-lattice-tools/shared/mathutil"
)

// Provenance describes where a parameter/value came from.
type Provenance string

const (
	ProvenanceMeasured    Provenance = "measured"
	ProvenanceCalibrated  Provenance = "calibrated"
	ProvenanceEstimated   Provenance = "estimated"
	ProvenancePlaceholder Provenance = "placeholder"
)

// ConfidenceTag tags a physics value with provenance and confidence [0,1].
type ConfidenceTag struct {
	Provenance Provenance `json:"provenance"`
	Confidence float64    `json:"confidence"`
}

// ConfidenceLedger stores parameter provenance metadata.
type ConfidenceLedger struct {
	registry map[string]ConfidenceTag
}

// NewConfidenceLedger builds a ledger with baseline known physics parameters.
func NewConfidenceLedger() *ConfidenceLedger {
	l := &ConfidenceLedger{registry: make(map[string]ConfidenceTag)}
	l.Register("Pr", ConfidenceTag{Provenance: ProvenanceMeasured, Confidence: 0.95})
	l.Register("Ps", ConfidenceTag{Provenance: ProvenanceMeasured, Confidence: 0.93})
	l.Register("Ec", ConfidenceTag{Provenance: ProvenanceMeasured, Confidence: 0.92})
	l.Register("beta_landau", ConfidenceTag{Provenance: ProvenanceCalibrated, Confidence: 0.86})
	l.Register("gamma_landau", ConfidenceTag{Provenance: ProvenanceCalibrated, Confidence: 0.86})
	l.Register("rho_viscosity", ConfidenceTag{Provenance: ProvenanceEstimated, Confidence: 0.72})
	return l
}

// Register stores/overwrites a parameter tag.
func (l *ConfidenceLedger) Register(parameter string, tag ConfidenceTag) {
	if l == nil {
		return
	}
	tag.Confidence = mathutil.Clamp01(tag.Confidence)
	l.registry[strings.ToLower(strings.TrimSpace(parameter))] = tag
}

// Lookup returns parameter provenance metadata.
func (l *ConfidenceLedger) Lookup(parameter string) (ConfidenceTag, bool) {
	if l == nil {
		return ConfidenceTag{}, false
	}
	tag, ok := l.registry[strings.ToLower(strings.TrimSpace(parameter))]
	return tag, ok
}

// TagOutput returns known metadata, or a placeholder tag if unknown.
func (l *ConfidenceLedger) TagOutput(parameter string, value float64) TaggedPhysicsValue {
	tag, ok := l.Lookup(parameter)
	if !ok {
		tag = ConfidenceTag{Provenance: ProvenancePlaceholder, Confidence: 0.1}
	}
	return TaggedPhysicsValue{Parameter: parameter, Value: value, Tag: tag}
}

// TaggedPhysicsValue represents a value plus confidence/provenance metadata.
type TaggedPhysicsValue struct {
	Parameter string        `json:"parameter"`
	Value     float64       `json:"value"`
	Tag       ConfidenceTag `json:"tag"`
}
