package eda

type EDAState struct {
	DesignName     string   `json:"design_name"`
	ProcessNode    string   `json:"process_node"`
	ArrayRows      int      `json:"array_rows"`
	ArrayCols      int      `json:"array_cols"`
	ExportFormats  []string `json:"export_formats"`
	TotalCells     int      `json:"total_cells"`
	AreaMM2        float64  `json:"area_mm2"`
	PowerMW        float64  `json:"power_mw"`
	SPICESnippet   string   `json:"spice_snippet"`
	VerilogSnippet string   `json:"verilog_snippet"`
	DEFSnippet     string   `json:"def_snippet"`
	LEFSnippet     string   `json:"lef_snippet"`
}
