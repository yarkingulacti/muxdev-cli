package portkill

// Process describes a process bound to a TCP port.
type Process struct {
	PID     int
	Command string
}
