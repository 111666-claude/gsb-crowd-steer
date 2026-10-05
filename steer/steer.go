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
	// 幂等：同一帧重复计算不重复计入检查数。
	f.Scanned = 0
	f.Deferred = 0

	// 输出顺序只由代理 ID 决定，与并发调度无关。
	order := make([]int, len(agents))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool {
		return agents[order[a]].ID < agents[order[b]].ID
	})

	// 网格分格索引：候选只取本格与相邻格，不两两配对。
	cellSize := f.Radius
	if cellSize < 1 {
		cellSize = 1
	}
	grid := make(map[cell][]int, len(agents))
	for i, a := range agents {
		key := cell{floorDiv(a.X, cellSize), floorDiv(a.Y, cellSize)}
		grid[key] = append(grid[key], i)
	}

	candidates := make([][]int, len(agents))
	for _, i := range order {
		key := cell{floorDiv(agents[i].X, cellSize), floorDiv(agents[i].Y, cellSize)}
		var cand []int
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				for _, j := range grid[cell{key.x + dx, key.y + dy}] {
					if j != i {
						cand = append(cand, j)
					}
				}
			}
		}
		candidates[i] = cand
	}

	// 检查预算：按 ID 顺序累计候选对数，超出的代理本帧延后。
	computed := make([]int, 0, len(agents))
	scanned := 0
	for _, i := range order {
		if scanned+len(candidates[i]) > f.MaxChecks {
			break
		}
		scanned += len(candidates[i])
		computed = append(computed, i)
	}
	f.Scanned = scanned
	f.Deferred = len(agents) - len(computed)

	results := make([]Result, len(computed))
	var wg sync.WaitGroup
	chunk := (len(computed) + f.Workers - 1) / f.Workers
	if chunk < 1 {
		chunk = 1
	}
	for w := 0; w < f.Workers; w++ {
		lo := w * chunk
		hi := lo + chunk
		if hi > len(computed) {
			hi = len(computed)
		}
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			for k := lo; k < hi; k++ {
				i := computed[k]
				me := agents[i]
				res := Result{ID: me.ID}
				for _, j := range candidates[i] {
					other := agents[j]
					dx := other.X - me.X
					dy := other.Y - me.Y
					if abs(dx) <= f.Radius && abs(dy) <= f.Radius {
						res.Neighbors++
						res.SumX += dx
						res.SumY += dy
					}
				}
				results[k] = res
			}
		}(lo, hi)
	}
	wg.Wait()

	return results
}

type cell struct{ x, y int }

// floorDiv 向下取整除法，保证负坐标也落在正确的格子里。
func floorDiv(v, d int) int {
	q := v / d
	if v%d != 0 && (v < 0) != (d < 0) {
		q--
	}
	return q
}
