package portkill

// LineHandler receives stdout/stderr lines from an attached process.
type LineHandler func(stderr bool, text string)
