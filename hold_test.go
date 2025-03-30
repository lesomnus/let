package let_test

import (
	"context"
	"testing"
	"time"

	"github.com/lesomnus/let"
)

func TestHold(t *testing.T) {
	t.Run("tasks are deferred", func(t *testing.T) {
		c := make(chan struct{})
		worker := let.NewWorker()
		worker = let.Hold(worker)
		worker.Go(let.New(func(ctx context.Context) error {
			<-c
			return nil
		}))

		select {
		case c <- struct{}{}:
			t.Fail()
		case <-time.After(100 * time.Millisecond):
		}
	})
	t.Run("tasks are stared on run", func(t *testing.T) {
		c := make(chan struct{})
		worker := let.NewWorker()
		worker = let.Hold(worker)
		worker.Go(let.New(func(ctx context.Context) error {
			<-c
			return nil
		}))
		go worker.Run(t.Context())

		select {
		case c <- struct{}{}:
		case <-time.After(100 * time.Millisecond):
			t.Fail()
		}
	})
}
