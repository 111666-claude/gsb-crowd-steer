// Package steer 为一群代理计算邻域避让。
package steer

import (
	"sort"
	"sync"
)

// Agent 是场上一个代理。
type Agent struct {
	ID   string
	X, Y int
}

// Result 是某个代理这一帧的避让结果。
type Result struct {
	ID        string
	Neighbors int
	SumX      int
	SumY      int
}

// Field 按帧计算避让。
type Field struct {
	Radius    int
	MaxChecks int
	Workers   int

	Scanned  int
	Deferred int
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Step 计算这一帧的结果，按代理 ID 升序返回。
func (f *Field) Step(agents []Agent) []Result {
	if f.Workers < 1 {
		f.Workers = 1
	}
	results := make([]Result, len(agents))
	checks := make([]int, len(agents))

	var wg sync.WaitGroup
	chunk := (len(agents) + f.Workers - 1) / f.Workers
	for w := 0; w < f.Workers; w++ {
		lo := w * chunk
		hi := lo + chunk
		if hi > len(agents) {
			hi = len(agents)
		}
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				me := agents[i]
				res := Result{ID: me.ID}
				for j := 0; j < len(agents); j++ {
					if j == i {
						continue
					}
					other := agents[j]
					dx := other.X - me.X
					dy := other.Y - me.Y
					checks[i]++
					if abs(dx) < f.Radius && abs(dy) < f.Radius {
						res.Neighbors++
						res.SumX += dx
						res.SumY += dy
					}
				}
				results[i] = res
			}
		}(lo, hi)
	}
	wg.Wait()

	for _, n := range checks {
		f.Scanned += n
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Neighbors > results[j].Neighbors
	})
	return results
}
