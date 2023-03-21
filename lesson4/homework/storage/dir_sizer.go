package storage

import (
	"context"
	"golang.org/x/sync/errgroup"
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

func (a *sizer) Size(ctx context.Context, d Dir) (Result, error) {
	// TODO: implement this
	errorQueue := new(errgroup.Group)
	dirQueue := make(chan Dir, QUEUE_SIZE)
	res := Result{}
	var tasksToDo int64 = 1
	dirQueue <- d

	cxtEmptyQueue, cancel := context.WithCancel(ctx)

	defer cancel()
	for i := 0; i < a.maxWorkersCount; i++ {
		errorQueue.Go(func() error {
			for {
				select {
				case <-cxtEmptyQueue.Done():
					return nil
				default:

					cur, ok := <-dirQueue
					if !ok {
						return nil
					}
					dirs, files, err := cur.Ls(ctx)
					if err != nil {
						close(dirQueue)
						return err
					}
					atomic.AddInt64(&tasksToDo, int64(len(dirs)))
					localRes := Result{}
					for _, f := range files {
						fileSize, err := f.Stat(ctx)
						if err != nil {
							close(dirQueue)
							return err
						}
						localRes.Size += fileSize
						localRes.Count++
					}
					atomic.AddInt64(&res.Size, localRes.Size)
					atomic.AddInt64(&res.Count, localRes.Count)
					for _, d := range dirs {
						dirQueue <- d
					}
					atomic.AddInt64(&tasksToDo, -1)
					if atomic.LoadInt64(&tasksToDo) == 0 {
						close(dirQueue)
						cancel()
					}
					continue
				}
			}
		})
	}
	err := errorQueue.Wait()
	return res, err
}
