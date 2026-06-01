package services

// command type of monitor manager service...
type MMCommand struct {
	Command       string  `json:"command"`
	TargetMonitor string  `json:"targetMonitor"`
	Value         float64 `json:"value"`
	Prefix        rune    `json:"prefix"`
}
