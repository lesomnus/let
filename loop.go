package let

import "context"

type loop struct {
	Task
}

// Loop creates a Task that repeatedly runs the given Task until it returns an error.
func Loop(ts ...Task) Task {
	return loop{Seq(ts...)}
}

func (t loop) Run(ctx context.Context) error {
	for {
		if err := t.Task.Run(ctx); err != nil {
			return err
		}
	}
}
