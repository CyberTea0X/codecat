package progress

import (
	"fmt"
	"os"
	"sync/atomic"
)

type Counter struct {
	count int64
	show  bool
}

func New() *Counter {
	return &Counter{}
}

func (c *Counter) Show() {
	c.show = true
}

func (c *Counter) Increment() {
	if !c.show {
		return
	}
	count := atomic.AddInt64(&c.count, 1)
	fmt.Fprintf(os.Stderr, "\rScanned %d files...", count)
}

func (c *Counter) Done() {
	if !c.show {
		return
	}
	fmt.Fprintf(os.Stderr, "\rScanned %d files total.\n", atomic.LoadInt64(&c.count))
}
