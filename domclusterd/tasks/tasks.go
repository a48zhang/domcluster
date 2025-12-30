package tasks

import (
	"context"
	"sync"

	"domclusterd/log"

	"go.uber.org/zap"
)

type Task interface {
	Priority() int
	Run(context.Context) error
}

const TASKMANAGER_SIZE = 64
const TASKWORKERPOOL_SIZE = 4

type TaskManager struct {
	tasks        chan Task
	deferedtasks chan Task
	workerPool   WorkerPool
	ctx          context.Context
}

func NewTaskManager(ctx context.Context) *TaskManager {
	return &TaskManager{
		tasks:        make(chan Task, TASKMANAGER_SIZE),
		deferedtasks: make(chan Task, TASKMANAGER_SIZE),
		workerPool:   *NewWorkerPool(ctx, TASKWORKERPOOL_SIZE),
		ctx:          ctx,
	}
}

func (tm *TaskManager) Add(task Task) error {
	if task.Priority() < 0 {
		select {
		case tm.deferedtasks <- task:
			return nil
		case <-tm.ctx.Done():
			return tm.ctx.Err()
		}
	} else if task.Priority() > 0 {
		go func() {
			if err := task.Run(tm.ctx); err != nil {
				log.Error("high priority task error", zap.Error(err))
			}
		}()
		return nil
	} else {
		select {
		case tm.tasks <- task:
			return nil
		case <-tm.ctx.Done():
			return tm.ctx.Err()
		}
	}
}

func (tm *TaskManager) Run() {
	go func() {
		for {
			select {
			case task, ok := <-tm.tasks:
				if !ok {
					return
				}
				tm.workerPool.Tasks <- task
			case <-tm.ctx.Done():
				return
			}
		}
	}()
}

func (tm *TaskManager) Stop() {
	wg := sync.WaitGroup{}
	for {
		ok := false
		select {
		case task := <-tm.deferedtasks:
			wg.Add(1)
			go func(t Task) {
				defer wg.Done()
				t.Run(tm.ctx)
			}(task)
		default:
			ok = true
		}
		if ok {
			break
		}
	}
	close(tm.tasks)
	tm.workerPool.Stop()
	close(tm.deferedtasks)
	wg.Wait()
	return
}
