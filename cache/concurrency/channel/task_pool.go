package channel

import "context"

type Task func()

type TaskPool struct {
	tasks chan Task
	close chan struct{}
}

func NewTaskPool(numG int, capacity int) *TaskPool {
	res := &TaskPool{
		tasks: make(chan Task, capacity),
		close: make(chan struct{}),
	}
	for i := 0; i < numG; i++ {
		go func() {
			select {
			case <-res.close:
				return
			case task := <-res.tasks:
				task()
			}
		}()
	}
	return res
}

func (tp *TaskPool) Submit(ctx context.Context, t Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case tp.tasks <- t:
	}
	return nil
}

func (tp *TaskPool) Close() {
	close(tp.tasks)
}
