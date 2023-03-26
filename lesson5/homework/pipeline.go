package executor

import (
	"context"
)

type (
	In  <-chan any
	Out = In
)

type Stage func(in In) (out Out)

//func switcher(in In) Out {
//  out := make(chan any)
//  go func() {
//    defer close(out)
//    for {
//      for v := range in {
//        out <- v
//      }
//    }
//  }()
//  return out
//}

func ExecutePipeline(ctx context.Context, in In, stages ...Stage) Out {
	//q := append([]Stage{switcher()}, stages...)
	outList := make([]Out, 0, len(stages)+1)
	outList = append(outList, in)
	for i := 0; i < len(stages); i++ {
		outList = append(outList, stages[i](outList[i]))
	}

	var res = make(chan any)

	go func() {
		defer close(res)

		for v := range outList[len(stages)] {
			select {
			case <-ctx.Done():
				return
			default:
				res <- v
			}
		}
	}()
	return res
}
