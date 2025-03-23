package let

import "context"

type loop struct {
	Task
}

// Loop creates a Task that repeatedly runs the given Task until it returns an error.
func Loop(t Task) Task {
	return loop{t}
}

func (t loop) Run(ctx context.Context) error {
	for {
		err := t.Task.Run(ctx)
		if err == nil {
			continue
		}

		return err
	}
}
