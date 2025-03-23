package let

import "sync/atomic"

// Close the channel once.
type closer struct {
	closed atomic.Bool
	done   chan struct{}
}

func initCloser(c *closer) {
	c.done = make(chan struct{})
}

func (c *closer) close() {
	// Ensure the channel to be closed only once.
	// Note that slow-path implemented by sync.Once does not needed here.
	if !c.closed.Swap(true) {
		close(c.done)
	}
}
