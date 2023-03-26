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
	outList := make([]Out, 0, len(stages)+1)
	outList = append(outList, in)
	for i := 0; i < len(stages); i++ {
		outList = append(outList, stages[i](outList[i]))
	}

	var res = make(chan any)

	go func() {
		defer close(res)

		for {
			select {
			case <-ctx.Done():
				return
			case d := <-outList[len(stages)]:
				if d == nil {
					return
				}
				res <- d
			}
		}
	}()
	return res
}
