package workers

import (
	"sync"
)

// T is a type alias to accept any type.
type T = interface{}

// WorkerPoolInterface is a contract for Worker Pool implementation
type WorkerPoolInterface interface {
	SetFunc(workerFunc interface{})
	Run()
	AddTask(task T)
	Stop()
}

type WorkerPool struct {
	maxWorker   int
	workerFunc  func(string) (string, error)
	queuedTaskC chan string
	ResultC     chan string
	ErrorC      chan error
	wg          sync.WaitGroup
}

func NewWorkerPool(maxWorker int, workerFunc func(string) (string, error)) *WorkerPool {
	taskC := make(chan string, maxWorker)
	resC := make(chan string, maxWorker)
	errC := make(chan error, maxWorker)
	return &WorkerPool{
		maxWorker:   maxWorker,
		workerFunc:  workerFunc,
		queuedTaskC: taskC,
		ResultC:     resC,
		ErrorC:      errC,
	}
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		wp.wg.Add(1)
		go func(workerID int) {
			defer wp.wg.Done()
			for task := range wp.queuedTaskC {
				res, err := wp.workerFunc(task)
				if err != nil {
					wp.ErrorC <- err
				} else {
					wp.ResultC <- res
				}
			}
		}(i + 1)
	}
}

func (wp *WorkerPool) AddTask(task string) {
	wp.queuedTaskC <- task
}

func (wp *WorkerPool) Stop() {
	close(wp.queuedTaskC)
	wp.wg.Wait()
	close(wp.ResultC)
	close(wp.ErrorC)
}
