package outlog

// Define time units accepted by Massif

type TimeUnit int

const (
	I TimeUnit = iota
	B
	MS
	AUTO
)
