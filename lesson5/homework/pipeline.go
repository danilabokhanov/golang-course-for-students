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

var switchChan chan any

func switcher(ctx context.Context) Stage {
	return func(in In) Out {
		go func() {
			defer close(switchChan)
			for v := range in {
				select {
				case <-ctx.Done():
					return
				default:
					switchChan <- v
				}
			}
		}()
		return switchChan
	}
}

func ExecutePipeline(ctx context.Context, in In, stages ...Stage) Out {
	//q := append([]Stage{switcher()}, stages...)
	switchChan = make(chan any)
	stages = append([]Stage{switcher(ctx)}, stages...)
	outList := make([]Out, 0, len(stages)+1)
	outList = append(outList, in)
	for i := 0; i < len(stages); i++ {
		outList = append(outList, stages[i](outList[i]))
	}

	var res = make(chan any)

	go func() {
		defer close(res)
		var storage = []any{}
		byContext := false
		for v := range outList[len(stages)] {
			select {
			case <-ctx.Done():
				byContext = true
				break
			default:
				storage = append(storage, v)
			}
		}
		if byContext {
			return
		}
		for _, v := range storage {
			res <- v
		}
	}()
	return res
}
