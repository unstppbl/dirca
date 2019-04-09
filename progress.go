package dirber

import (
	"context"
	"fmt"
	"os"
	"time"
)

// PrintProgress outputs the current wordlist progress to stderr
func (d *Dirber) printProgress(requestsIssued *int) {
	d.mu.RLock()
	if d.opts.wordlistLength > 0 {
		fmt.Fprintf(os.Stderr, "\rProgress: %d / %d (%3.2f%%)", *requestsIssued, d.opts.wordlistLength, float32(*requestsIssued)*100.0/float32(d.opts.wordlistLength))
	}
	d.mu.RUnlock()
}

// ClearProgress removes the last status line from stderr
func (d *Dirber) ClearProgress() {
	fmt.Fprint(os.Stderr, "\r\r")
}

func (d *Dirber) progressWorker(c context.Context, requestsIssued *int) {
	tick := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-tick.C:
			d.printProgress(requestsIssued)
		case <-c.Done():
			return
		}
	}
}

func (d *Dirber) incrementRequests(requestsIssued *int) {
	d.mu.Lock()
	*requestsIssued++
	d.mu.Unlock()
}
