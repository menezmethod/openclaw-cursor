package streaming

// DeltaTracker tracks cumulative text and emits incremental deltas.
type DeltaTracker struct {
	lastText    string
	lastThinking string
}

// NextText returns the delta for new text and updates state.
func (d *DeltaTracker) NextText(value string) string {
	delta := diff(d.lastText, value)
	d.lastText = value
	return delta
}

// NextThinking returns the delta for new thinking content and updates state.
func (d *DeltaTracker) NextThinking(value string) string {
	delta := diff(d.lastThinking, value)
	d.lastThinking = value
	return delta
}

// Reset clears the tracker state.
func (d *DeltaTracker) Reset() {
	d.lastText = ""
	d.lastThinking = ""
}

func diff(previous, current string) string {
	if previous == "" {
		return current
	}
	// current should be previous + new suffix
	if len(current) >= len(previous) && current[:len(previous)] == previous {
		return current[len(previous):]
	}
	return current
}
