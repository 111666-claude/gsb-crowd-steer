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
	f.Scanned = 0
	f.Deferred = 0

	cell := f.Radius
	if cell < 1 {
		cell = 1
	}

	order := make([]int, len(agents))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool {
		return agents[order[a]].ID < agents[order[b]].ID
	})

	type point struct{ x, y int }
	grid := make(map[point][]int)
	for _, i := range order {
		c := point{floorDiv(agents[i].X, cell), floorDiv(agents[i].Y, cell)}
		grid[c] = append(grid[c], i)
	}

	candidates := make([][]int, len(agents))
	for pos, i := range order {
		cx := floorDiv(agents[i].X, cell)
		cy := floorDiv(agents[i].Y, cell)
		var list []int
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				for _, j := range grid[point{cx + dx, cy + dy}] {
					if j != i {
						list = append(list, j)
					}
				}
			}
		}
		candidates[pos] = list
	}

	used := 0
	computed := len(agents)
	for pos := range order {
		need := len(candidates[pos])
		if used+need > f.MaxChecks {
			computed = pos
			break
		}
		used += need
	}
	f.Scanned = used
	f.Deferred = len(agents) - computed

	results := make([]Result, computed)
	var wg sync.WaitGroup
	chunk := (computed + f.Workers - 1) / f.Workers
	for w := 0; w < f.Workers; w++ {
		lo := w * chunk
		hi := lo + chunk
		if hi > computed {
			hi = computed
		}
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			for pos := lo; pos < hi; pos++ {
				me := agents[order[pos]]
				res := Result{ID: me.ID}
				for _, j := range candidates[pos] {
					other := agents[j]
					dx := other.X - me.X
					dy := other.Y - me.Y
					if abs(dx) <= f.Radius && abs(dy) <= f.Radius {
						res.Neighbors++
						res.SumX += dx
						res.SumY += dy
					}
				}
				results[pos] = res
			}
		}(lo, hi)
	}
	wg.Wait()
	return results
}

func floorDiv(v, s int) int {
	q := v / s
	if v < 0 && v%s != 0 {
		q--
	}
	return q
}
