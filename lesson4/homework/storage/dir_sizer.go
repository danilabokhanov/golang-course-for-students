package storage

import (
	"context"
	"golang.org/x/sync/errgroup"
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

func (a *sizer) Size(ctx context.Context, d Dir) (Result, error) {
	// TODO: implement this
	errorQueue := new(errgroup.Group)
	dirQueue := []Dir{d}
	isOpen := atomic.Value{}
	isOpen.Store(true)
	res := Result{}
	var tasksToDo int64 = 1
	mt := sync.Mutex{}
	cxtEmptyQueue, cancel := context.WithCancel(ctx)

	defer cancel()
	for i := 0; i < a.maxWorkersCount; i++ {
		errorQueue.Go(func() error {
			for {
				select {
				case <-cxtEmptyQueue.Done():
					return nil
				default:
					if !isOpen.Load().(bool) {
						return nil
					}
					mt.Lock()
					if len(dirQueue) == 0 {
						mt.Unlock()
						continue
					}
					cur := dirQueue[len(dirQueue)-1]
					dirQueue = dirQueue[:len(dirQueue)-1]
					mt.Unlock()
					dirs, files, err := cur.Ls(ctx)
					if err != nil {
						isOpen.Store(false)
						return err
					}
					atomic.AddInt64(&tasksToDo, int64(len(dirs)))
					localRes := Result{}
					for _, f := range files {
						fileSize, err := f.Stat(ctx)
						if err != nil {
							isOpen.Store(false)
							return err
						}
						localRes.Size += fileSize
						localRes.Count++
					}
					atomic.AddInt64(&res.Size, localRes.Size)
					atomic.AddInt64(&res.Count, localRes.Count)
					localDirs := []Dir{}
					for _, d := range dirs {
						localDirs = append(localDirs, d)
					}
					mt.Lock()
					dirQueue = append(dirQueue, localDirs...)
					mt.Unlock()
					atomic.AddInt64(&tasksToDo, -1)
					if atomic.LoadInt64(&tasksToDo) == 0 {
						isOpen.Store(false)
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
