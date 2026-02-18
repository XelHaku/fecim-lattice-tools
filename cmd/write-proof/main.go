package main

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	sharedtheme "fecim-lattice-tools/shared/theme"
)

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func main() {
	outDir := "/tmp/write_proof"
	os.MkdirAll(outDir, 0755)

	a := app.NewWithID("com.fecim.write.proof")
	a.Settings().SetTheme(&sharedtheme.FeCIMTheme{})
	w := a.NewWindow("Module 4 - Write Cell Proof")
	w.Resize(fyne.NewSize(1400, 900))

	emb := gui.NewEmbeddedCircuitsApp()
	done := make(chan struct{})

	a.Lifecycle().SetOnStarted(func() {
		content := emb.BuildContent(a, w)
		bg := canvas.NewRectangle(theme.BackgroundColor())
		w.SetContent(container.NewMax(bg, content))
		w.Show()
		emb.Start()

		time.AfterFunc(2*time.Second, func() {
			fyne.DoAndWait(func() {
				// ---- BEFORE WRITE ----
				w.Canvas().Refresh(w.Content())
				time.Sleep(200 * time.Millisecond)
				img := w.Canvas().Capture()
				savePNG(outDir+"/01_before_write.png", img)
				fmt.Println("[proof] 01_before_write.png captured")
			})

			// Perform ISPP write via DeviceState
			ds := emb.GetDeviceState()

			// Select cell (2,2), target level 20
			ds.SetSelectedCell(2, 2)
			ds.SetOperationMode(gui.OpModeWrite)

			// Read current level from ISPP status snapshot.
			levelBefore := ds.GetISPPStatus().CurrentLevel
			fmt.Printf("[proof] Cell(2,2) BEFORE write (status snapshot): level=%d\n", levelBefore)

			// Run ISPP write loop (API returns enum status; details are in GetISPPStatus())
			currentLevel := levelBefore
			targetLevel := 20
			ds.StartISPP(2, 2, targetLevel, currentLevel)
			for i := 0; i < 50; i++ {
				result := ds.ISPPIterate(currentLevel)
				status := ds.GetISPPStatus()
				currentLevel = status.CurrentLevel
				if result == gui.ISPPResultVerified || result == gui.ISPPResultMaxIterations || result == gui.ISPPResultNotActive {
					fmt.Printf("[proof] ISPP done: result=%v level=%d iter=%d\n", result, currentLevel, status.Iteration)
					break
				}
			}

			levelAfter := ds.GetISPPStatus().CurrentLevel
			fmt.Printf("[proof] Cell(2,2) AFTER write (status snapshot): level=%d (target=%d)\n", levelAfter, targetLevel)

			// ---- AFTER WRITE ----
			fyne.DoAndWait(func() {
				w.Canvas().Refresh(w.Content())
				time.Sleep(300 * time.Millisecond)
				img := w.Canvas().Capture()
				savePNG(outDir+"/02_after_write.png", img)
				fmt.Println("[proof] 02_after_write.png captured")

				// Write result to text file
				err := os.WriteFile(outDir+"/result.txt",
					[]byte(fmt.Sprintf("Cell(2,2): before=%d after=%d target=20 success=%v\n",
						levelBefore, levelAfter, math.Abs(float64(levelAfter-20)) <= 1)),
					0644)
				if err != nil {
					fmt.Println("[proof] write result.txt failed:", err)
				}
			})

			emb.Stop()
			w.Close()
			close(done)
			a.Quit()
		})
	})

	a.Run()
	<-done
	fmt.Println("[proof] Complete. Images in", outDir)
}
