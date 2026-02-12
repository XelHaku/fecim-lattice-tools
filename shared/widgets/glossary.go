// Package widgets provides reusable UI components.
package widgets

import (
	"net/url"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// GlossaryEntry represents a single glossary term.
type GlossaryEntry struct {
	Term       string
	Definition string
	Category   string // Physics, Architecture, Circuits, Metrics
}

// TermsData contains all technical terms used across modules.
var TermsData = []GlossaryEntry{
	// Physics
	{
		Term:       "FeCIM",
		Definition: "Ferroelectric Compute-in-Memory - analog computation using ferroelectric memory cells with programmable polarization states. Enables matrix-vector multiplication in hardware using Ohm's law (V=IR).",
		Category:   "Physics",
	},
	{
		Term:       "Ec",
		Definition: "Coercive Field - Electric field required to switch ferroelectric polarization state. Typical range: 0.6-1.5 MV/cm for HZO superlattices. Lower Ec reduces write energy but may compromise retention.",
		Category:   "Physics",
	},
	{
		Term:       "Pr",
		Definition: "Remnant Polarization - Polarization remaining after electric field removal. Reported literature ranges exist, but this simulator treats specific numeric values as modeled unless explicitly marked verified in HONESTY_AUDIT.md.",
		Category:   "Physics",
	},
	{
		Term:       "HZO",
		Definition: "Hafnium Zirconium Oxide (Hf₁₋ₓZrₓO₂) - ferroelectric superlattice material. CMOS-compatible. Conference claim: 30 discrete analog states per cell (~4.9 bits/cell), pending peer review; peer-reviewed devices report 32–140 states in related materials. Optimal composition: x ≈ 0.5.",
		Category:   "Physics",
	},
	{
		Term:       "Hysteresis Loop",
		Definition: "P-E (Polarization-Electric Field) curve showing path-dependent behavior. Area enclosed represents energy dissipation per cycle. Shape determines analog state density and retention.",
		Category:   "Physics",
	},
	{
		Term:       "Preisach Model",
		Definition: "Mathematical model for hysteresis using distribution of elementary switching operators (hysteron). Captures memory effects and minor loop behavior. Key parameters: α (coercive field), β (interaction field).",
		Category:   "Physics",
	},
	{
		Term:       "Endurance",
		Definition: "Number of write/erase cycles before device failure. Literature reports wide ranges; in this UI, treat numbers as reported context unless verified in HONESTY_AUDIT.md.",
		Category:   "Physics",
	},

	// Architecture
	{
		Term:       "1T1R",
		Definition: "One Transistor, One Resistor architecture. Transistor acts as selector to eliminate sneak paths in crossbar arrays. Enables high density (4F² footprint) with precise cell addressing.",
		Category:   "Architecture",
	},
	{
		Term:       "MVM",
		Definition: "Matrix-Vector Multiplication - Parallel computation in crossbar: I = G·V where I is output current vector, G is conductance matrix (programmed weights), V is input voltage vector. Single-step MAC operation via Kirchhoff's current law.",
		Category:   "Architecture",
	},
	{
		Term:       "MAC",
		Definition: "Multiply-Accumulate - Fundamental neural network operation: y += w·x. Traditional digital requires separate multiply and accumulate steps. FeCIM performs MAC in single analog step using Ohm's law.",
		Category:   "Architecture",
	},
	{
		Term:       "BEOL",
		Definition: "Back-End-Of-Line - Later stages of chip manufacturing (metal interconnect layers above transistors). FeFET integration demonstrated at 22nm BEOL (CEA-Leti 2024) enables 3D memory stacking without disturbing CMOS logic.",
		Category:   "Architecture",
	},
	{
		Term:       "Sneak Path",
		Definition: "Unintended current flow through neighboring cells in passive crossbar arrays. Causes read/write interference. Mitigated by 1T1R architecture (selector transistor per cell) or high selector nonlinearity.",
		Category:   "Architecture",
	},
	{
		Term:       "IR Drop",
		Definition: "Voltage loss across interconnect resistance in large crossbar arrays. Causes location-dependent read/write errors. Scales as O(N²) for N×N arrays. Mitigated by segmentation or peripheral driver circuits.",
		Category:   "Architecture",
	},

	// Circuits
	{
		Term:       "DAC",
		Definition: "Digital-to-Analog Converter - Converts digital input vectors to analog voltages for crossbar wordlines. Precision determines input quantization noise. Typical: 4-8 bits for neural network inference.",
		Category:   "Circuits",
	},
	{
		Term:       "ADC",
		Definition: "Analog-to-Digital Converter - Converts analog output currents from crossbar bitlines to digital values. SAR (successive approximation) ADCs common for low power. Resolution: 6-12 bits typical.",
		Category:   "Circuits",
	},
	{
		Term:       "TIA",
		Definition: "Transimpedance Amplifier - Converts crossbar output current to voltage with gain R_f: V_out = -I_in·R_f. Critical for low-noise readout. Bandwidth and offset current determine inference speed/accuracy.",
		Category:   "Circuits",
	},
	{
		Term:       "Sense Amplifier",
		Definition: "High-gain differential amplifier for detecting small signal differences in memory cells. Enables faster read operations but may limit analog state resolution (typically 3-5 bits).",
		Category:   "Circuits",
	},

	// Metrics
	{
		Term:       "TRL",
		Definition: "Technology Readiness Level - Scale from 1 (basic principles) to 9 (production ready). FeCIM status (Tour COSM 2025): TRL 4 (lab validation). Other ferroelectric memory technologies may be at higher TRLs; treat those as separate contexts.",
		Category:   "Metrics",
	},
	{
		Term:       "TOPS/W",
		Definition: "Tera-Operations Per Second per Watt - energy efficiency metric. Any numeric comparisons shown here are literature-reported context and are not simulator-verified unless listed in HONESTY_AUDIT.md.",
		Category:   "Metrics",
	},
	{
		Term:       "Bits per Cell",
		Definition: "Information density in single memory element. FeCIM demo baseline: ~4.9 bits/cell (30 analog states, conference claim), up to 6.1-7.1 bits/cell (140 states demonstrated by Song 2024). NAND flash: 2-4 bits/cell.",
		Category:   "Metrics",
	},
	{
		Term:       "MNIST Accuracy",
		Definition: "Handwritten digit classification accuracy (0-9). State-of-art FeCIM: 98.24% (FTJ reservoir computing, ScienceDirect 2025), 96.6% (HZO crossbar, Nature Commun. 2023). Software: 99.7%.",
		Category:   "Metrics",
	},
	{
		Term:       "Retention Time",
		Definition: "Duration data remains stored without refresh. HZO FeFETs: >10 years at 85°C (industry standard). Improves at cryogenic temperatures (5K operation demonstrated). Critical for non-volatile applications.",
		Category:   "Metrics",
	},
	{
		Term:       "Write Energy",
		Definition: "Energy per bit for programming memory state. FeCIM: ~1-10 fJ/bit (ferroelectric switching). NAND flash: ~100 fJ/bit. Lower energy from smaller capacitance and voltage scaling (BEOL integration).",
		Category:   "Metrics",
	},
}

// GlossaryWidget displays searchable glossary with expandable terms.
type GlossaryWidget struct {
	widget.BaseWidget

	entries      []GlossaryEntry
	filteredList *widget.List
	searchEntry  *widget.Entry
	categoryTabs *container.AppTabs

	// Current state
	activeCategory string
	searchQuery    string
	filteredTerms  []GlossaryEntry
}

// NewGlossaryWidget creates a new glossary widget.
func NewGlossaryWidget() *GlossaryWidget {
	gw := &GlossaryWidget{
		entries:        TermsData,
		activeCategory: "All",
		filteredTerms:  TermsData,
	}

	gw.searchEntry = widget.NewEntry()
	gw.searchEntry.SetPlaceHolder("Search terms...")
	gw.searchEntry.OnChanged = func(query string) {
		gw.searchQuery = query
		gw.updateFilteredTerms()
	}

	gw.filteredList = widget.NewList(
		func() int {
			return len(gw.filteredTerms)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewRichTextFromMarkdown("**Term**"),
				widget.NewLabel("Definition..."),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(gw.filteredTerms) {
				return
			}
			entry := gw.filteredTerms[id]
			vbox := obj.(*fyne.Container)

			termLabel := widget.NewRichTextFromMarkdown("**" + entry.Term + "**")
			defLabel := widget.NewLabel(entry.Definition)
			defLabel.Wrapping = fyne.TextWrapWord

			vbox.Objects[0] = termLabel
			vbox.Objects[1] = defLabel
		},
	)

	gw.categoryTabs = container.NewAppTabs(
		container.NewTabItem("All", gw.createCategoryContent("All")),
		container.NewTabItem("Physics", gw.createCategoryContent("Physics")),
		container.NewTabItem("Architecture", gw.createCategoryContent("Architecture")),
		container.NewTabItem("Circuits", gw.createCategoryContent("Circuits")),
		container.NewTabItem("Metrics", gw.createCategoryContent("Metrics")),
	)
	gw.categoryTabs.OnSelected = func(tab *container.TabItem) {
		gw.activeCategory = tab.Text
		gw.updateFilteredTerms()
	}

	gw.ExtendBaseWidget(gw)
	return gw
}

// createCategoryContent creates content for a category tab.
func (gw *GlossaryWidget) createCategoryContent(category string) fyne.CanvasObject {
	// Placeholder - actual content rendered via filteredList
	return widget.NewLabel("")
}

// updateFilteredTerms filters entries based on category and search query.
func (gw *GlossaryWidget) updateFilteredTerms() {
	filtered := []GlossaryEntry{}

	query := strings.ToLower(gw.searchQuery)

	for _, entry := range gw.entries {
		// Filter by category
		if gw.activeCategory != "All" && entry.Category != gw.activeCategory {
			continue
		}

		// Filter by search query
		if query != "" {
			termMatch := strings.Contains(strings.ToLower(entry.Term), query)
			defMatch := strings.Contains(strings.ToLower(entry.Definition), query)
			if !termMatch && !defMatch {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	// Sort alphabetically by term
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Term < filtered[j].Term
	})

	gw.filteredTerms = filtered
	gw.filteredList.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (gw *GlossaryWidget) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Technical Glossary", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			gw.searchEntry,
			gw.categoryTabs,
		),
		nil, nil, nil,
		gw.filteredList,
	)

	return widget.NewSimpleRenderer(content)
}

// ShowGlossary displays a popup dialog with term definition.
func ShowGlossary(term string, parent fyne.Window) {
	var found *GlossaryEntry
	termLower := strings.ToLower(term)

	for i := range TermsData {
		if strings.ToLower(TermsData[i].Term) == termLower {
			found = &TermsData[i]
			break
		}
	}

	if found == nil {
		dialog.ShowInformation("Term Not Found",
			"No glossary entry found for: "+term,
			parent)
		return
	}

	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("**"+found.Term+"**"),
		widget.NewSeparator(),
		widget.NewLabel(found.Definition),
		layout.NewSpacer(),
		widget.NewLabel("Category: "+found.Category),
	)

	d := dialog.NewCustom(found.Term, "Close", content, parent)
	d.Resize(fyne.NewSize(500, 250))
	d.Show()
}

// ShowFullGlossary displays the complete glossary widget in a dialog.
func ShowFullGlossary(parent fyne.Window) {
	glossary := NewGlossaryWidget()

	d := dialog.NewCustom("Technical Glossary", "Close", glossary, parent)
	d.Resize(fyne.NewSize(700, 600))
	d.Show()
}

// ReferenceEntry represents a key scientific paper or document.
type ReferenceEntry struct {
	Title    string
	Citation string
	DOI      string // Empty if not applicable
	URL      string // Can be DOI link or documentation link
}

// ReferencesData contains key papers and documentation.
var ReferencesData = []ReferenceEntry{
	{
		Title:    "FeFET-Based Crossbar Array for MNIST Classification",
		Citation: "Nature Communications (2023)",
		DOI:      "10.1038/s41467-023-xxxxx",
		URL:      "https://doi.org/10.1038/s41467-023-xxxxx",
	},
	{
		Title:    "98.24% MNIST Accuracy with FTJ Reservoir Computing",
		Citation: "ScienceDirect (2025)",
		DOI:      "10.1016/j.xxxxx.2025.xxxxx",
		URL:      "https://www.sciencedirect.com/science/article/xxxxx",
	},
	{
		Title:    "10¹² Cycle Endurance with V:HfO₂ Doping",
		Citation: "Nano Letters (2024)",
		DOI:      "10.1021/acs.nanolett.xxxxx",
		URL:      "https://doi.org/10.1021/acs.nanolett.xxxxx",
	},
	{
		Title:    "512-Layer 3D FeFET Integration",
		Citation: "Nature (2025)",
		DOI:      "10.1038/s41586-025-xxxxx",
		URL:      "https://doi.org/10.1038/s41586-025-xxxxx",
	},
	{
		Title:    "22nm BEOL FeFET Demonstration",
		Citation: "CEA-Leti Technical Report (December 2024)",
		DOI:      "",
		URL:      "https://www.leti-cea.com/cea-tech/leti/english/Pages/xxxxx",
	},
	{
		Title:    "Automotive Grade FeFET (AEC-Q100)",
		Citation: "Fraunhofer IPMS (2024)",
		DOI:      "",
		URL:      "https://www.ipms.fraunhofer.de/xxxxx",
	},
	{
		Title:    "Cryogenic FeFET Operation (5K-300K)",
		Citation: "IEEE Transactions (2024)",
		DOI:      "10.1109/TED.2024.xxxxx",
		URL:      "https://doi.org/10.1109/TED.2024.xxxxx",
	},
	{
		Title:    "Scientific Honesty Audit",
		Citation: "Local Documentation",
		DOI:      "",
		URL:      "/docs/comparison/HONESTY_AUDIT.md",
	},
	{
		Title:    "Dr. Tour COSM 2025 Transcript",
		Citation: "Conference Transcript (Unverified)",
		DOI:      "",
		URL:      "/docs/videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md",
	},
	{
		Title:    "FeCIM Lattice Tools Repository",
		Citation: "Open Source Project (GitHub)",
		DOI:      "",
		URL:      "https://github.com/your-org/fecim-lattice-tools",
	},
}

// ReferencesWidget displays key papers with clickable DOI links.
type ReferencesWidget struct {
	widget.BaseWidget

	refs []ReferenceEntry
	list *widget.List
}

// NewReferencesWidget creates a new references widget.
func NewReferencesWidget() *ReferencesWidget {
	rw := &ReferencesWidget{
		refs: ReferencesData,
	}

	rw.list = widget.NewList(
		func() int {
			return len(rw.refs)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewRichTextFromMarkdown("**Title**"),
				widget.NewLabel("Citation"),
				widget.NewHyperlink("DOI/Link", nil),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(rw.refs) {
				return
			}
			ref := rw.refs[id]
			vbox := obj.(*fyne.Container)

			titleLabel := widget.NewRichTextFromMarkdown("**" + ref.Title + "**")
			citationLabel := widget.NewLabel(ref.Citation)
			citationLabel.Wrapping = fyne.TextWrapWord

			var linkWidget *widget.Hyperlink
			if ref.URL != "" {
				if strings.HasPrefix(ref.URL, "/") {
					// Local documentation link - create a label instead
					linkWidget = widget.NewHyperlink("(Local documentation)", nil)
					linkWidget.OnTapped = func() {
						// Could implement file opening here
					}
				} else {
					parsedURL, err := url.Parse(ref.URL)
					if err == nil {
						linkLabel := "View Paper"
						if ref.DOI != "" {
							linkLabel = "DOI: " + ref.DOI
						}
						linkWidget = widget.NewHyperlink(linkLabel, parsedURL)
					} else {
						linkWidget = widget.NewHyperlink("(Invalid URL)", nil)
					}
				}
			}

			if linkWidget == nil {
				linkWidget = widget.NewHyperlink("(No link available)", nil)
			}

			vbox.Objects[0] = titleLabel
			vbox.Objects[1] = citationLabel
			vbox.Objects[2] = linkWidget
		},
	)

	rw.ExtendBaseWidget(rw)
	return rw
}

// CreateRenderer implements fyne.Widget.
func (rw *ReferencesWidget) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Key References", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("*See HONESTY_AUDIT.md for verification status*"),
		),
		nil, nil,
		rw.list,
	)

	return widget.NewSimpleRenderer(content)
}

// ShowReferences displays the references widget in a dialog.
func ShowReferences(parent fyne.Window) {
	refs := NewReferencesWidget()

	d := dialog.NewCustom("Key References", "Close", refs, parent)
	d.Resize(fyne.NewSize(700, 500))
	d.Show()
}

// CreateHelpMenuItems creates standardized Help menu items for glossary and references.
func CreateHelpMenuItems(parent fyne.Window) []*fyne.MenuItem {
	return []*fyne.MenuItem{
		fyne.NewMenuItem("About the Science", func() {
			ShowAboutScience(parent)
		}),
		fyne.NewMenuItem("Technical Glossary", func() {
			ShowFullGlossary(parent)
		}),
		fyne.NewMenuItem("Key References", func() {
			ShowReferences(parent)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("About", func() {
			content := container.NewVBox(
				widget.NewLabelWithStyle("FeCIM Lattice Tools", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Ferroelectric Compute-in-Memory Simulation Suite"),
				widget.NewSeparator(),
				widget.NewLabel("Based on Dr. external research group's HfO₂-ZrO₂ superlattice research"),
				widget.NewLabel("30 discrete analog states per cell (~4.9 bits/cell, conference claim)"),
				layout.NewSpacer(),
				newGitHubLink(),
			)
			dialog.NewCustom("About", "Close", content, parent).Show()
		}),
	}
}

// newGitHubLink creates a hyperlink to the project's GitHub repository.
func newGitHubLink() *widget.Hyperlink {
	parsedURL, _ := url.Parse("https://github.com/your-org/fecim-lattice-tools")
	return widget.NewHyperlink("View on GitHub", parsedURL)
}

// QuickTermLookup returns definition for a term (case-insensitive).
// Returns empty string if not found.
func QuickTermLookup(term string) string {
	termLower := strings.ToLower(term)
	for _, entry := range TermsData {
		if strings.ToLower(entry.Term) == termLower {
			return entry.Definition
		}
	}
	return ""
}

// GetTermsByCategory returns all terms in a specific category.
func GetTermsByCategory(category string) []GlossaryEntry {
	filtered := []GlossaryEntry{}
	for _, entry := range TermsData {
		if entry.Category == category {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GetCategories returns all unique categories.
func GetCategories() []string {
	categories := make(map[string]bool)
	for _, entry := range TermsData {
		categories[entry.Category] = true
	}

	result := []string{}
	for cat := range categories {
		result = append(result, cat)
	}
	sort.Strings(result)
	return result
}

// CreateGlossaryButton creates a standardized button with glossary icon.
func CreateGlossaryButton(parent fyne.Window) *widget.Button {
	btn := widget.NewButtonWithIcon("Glossary", theme.InfoIcon(), func() {
		ShowFullGlossary(parent)
	})
	return btn
}

// CreateReferencesButton creates a standardized button for references.
func CreateReferencesButton(parent fyne.Window) *widget.Button {
	btn := widget.NewButtonWithIcon("References", theme.DocumentIcon(), func() {
		ShowReferences(parent)
	})
	return btn
}
