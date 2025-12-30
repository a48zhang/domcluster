package tasks

import (
	"context"
	"domclusterd/log"

	"go.uber.org/zap"
)

type WorkerPool struct {
	size  int
	Tasks chan Task
}

func NewWorkerPool(ctx context.Context, s int) *WorkerPool {
	c := make(chan Task, s*2)
	for range s {
		go func() {
			for {
				select {
				case t, ok := <-c:
					if !ok {
						return
					}
					err := t.Run(ctx)
					if err != nil {
						log.Error("task run error", zap.Error(err))
					}
				case <- ctx.Done():
					return
				}
			}
		}()
	}
	return &WorkerPool{
		size:  s,
		Tasks: c,
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.Tasks)
}
