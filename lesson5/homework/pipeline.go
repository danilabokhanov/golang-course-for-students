package executor

import (
	"context"
)

type (
	In  <-chan any
	Out = In
)

type Stage func(in In) (out Out)

func ExecutePipeline(ctx context.Context, in In, stages ...Stage) Out {
	out := in
	for i := 0; i < len(stages); i++ {
		out = stages[i](out)
	}

	var res = make(chan any)

	go func() {
		defer close(res)

		for {
			select {
			case <-ctx.Done():
				return
			case d := <-out:
				if d == nil {
					return
				}
				res <- d
			}
		}
	}()
	return res
}
