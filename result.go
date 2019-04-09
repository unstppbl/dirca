package dirber

// Result represents a single gobuster result
type Result struct {
	Entity string `json:"entity"`
	Status int    `json:"status"`
	Extra  string `json:"extra"`
	Size   int64  `json:"size"`
}
