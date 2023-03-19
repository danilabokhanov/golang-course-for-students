package storage

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

const MAX_GOROUTINES = 3
const QUEUE_SIZE = 100

// Result represents the Size function result
type Result struct {
	// Total Size of File objects
	Size int64
	// Count is a count of File objects processed
	Count int64
}

type DirSizer interface {
	// Size calculate a size of given Dir, receive a ctx and the root Dir instance
	// will return Result or error if happened
	Size(ctx context.Context, d Dir) (Result, error)
}

// sizer implement the DirSizer interface
type sizer struct {
	// maxWorkersCount number of workers for asynchronous run
	maxWorkersCount int

	// TODO: add other fields as you wish
}

// NewSizer returns new DirSizer instance
func NewSizer() DirSizer {
	return &sizer{MAX_GOROUTINES}
}

func Worker(ctx context.Context, errorQueue chan error, dirQueue chan Dir, globalRes *Result,
	tasksToDo *int64, mutex *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(errorQueue) == 1 {
				return
			}
			mutex.Lock()
			if *tasksToDo == 0 {
				if len(errorQueue) == 0 {
					errorQueue <- nil
				}
				mutex.Unlock()
				return
			}
			mutex.Unlock()
			cur, ok := <-dirQueue
			if !ok {
				return
			}
			dirs, files, err := cur.Ls(ctx)
			if err != nil && len(errorQueue) == 0 {
				errorQueue <- err
				return
			}
			atomic.AddInt64(tasksToDo, int64(len(dirs)))
			res := Result{}
			for _, f := range files {
				fileSize, err := f.Stat(ctx)
				if err != nil && len(errorQueue) == 0 {
					errorQueue <- err
					return
				}
				res.Size += fileSize
				res.Count++
			}
			atomic.AddInt64(&globalRes.Size, res.Size)
			atomic.AddInt64(&globalRes.Count, res.Count)
			for _, d := range dirs {
				dirQueue <- d
			}
			atomic.AddInt64(tasksToDo, -1)
			continue
		}
	}
}

func (a *sizer) Size(ctx context.Context, d Dir) (Result, error) {
	// TODO: implement this
	errorQueue := make(chan error, 1)
	dirQueue := make(chan Dir, QUEUE_SIZE)
	res := Result{}
	var tasksToDo int64 = 1
	var mutex sync.Mutex = sync.Mutex{}
	dirQueue <- d

	for i := 0; i < a.maxWorkersCount; i++ {
		go Worker(ctx, errorQueue, dirQueue, &res, &tasksToDo, &mutex)
	}

	err := <-errorQueue
	close(errorQueue)
	close(dirQueue)
	fmt.Print(res)
	return res, err
}
