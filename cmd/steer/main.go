// Command steer 打印一帧的邻域避让结果。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/crowd-steer/steer"
)

type line struct {
	ID        string `json:"id"`
	Neighbors int    `json:"neighbors"`
	SumX      int    `json:"sumx"`
	SumY      int    `json:"sumy"`
}

func gridAgents(cols, rows int) []steer.Agent {
	out := make([]steer.Agent, 0, cols*rows)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			out = append(out, steer.Agent{ID: fmt.Sprintf("a%03d_%03d", x, y), X: x * 2, Y: y * 2})
		}
	}
	return out
}

func scenario(name string) (*steer.Field, []steer.Agent) {
	switch name {
	case "boundary":
		return &steer.Field{Radius: 3, MaxChecks: 1000, Workers: 2}, []steer.Agent{
			{ID: "a", X: 0, Y: 0},
			{ID: "b", X: 3, Y: 0},
		}
	case "order":
		return &steer.Field{Radius: 5, MaxChecks: 1000, Workers: 2}, []steer.Agent{
			{ID: "a", X: 0, Y: 0},
			{ID: "b", X: 10, Y: 0},
			{ID: "c", X: 1, Y: 0},
			{ID: "d", X: 2, Y: 0},
		}
	case "budget":
		return &steer.Field{Radius: 5, MaxChecks: 20, Workers: 2}, gridAgents(4, 4)
	case "work":
		return &steer.Field{Radius: 1, MaxChecks: 10000000, Workers: 4}, gridAgents(50, 20)
	}
	fmt.Fprintf(os.Stderr, "未知场景 %q\n", name)
	os.Exit(2)
	return nil, nil
}

func main() {
	sample := flag.String("sample", "", "场景名")
	flag.Parse()

	if *sample == "" {
		fmt.Fprintln(os.Stderr, "需要 --sample=boundary|order|budget|work")
		os.Exit(2)
	}

	f, agents := scenario(*sample)
	results := f.Step(agents)

	if *sample == "work" {
		out, _ := json.Marshal(map[string]int{"agents": len(agents), "deferred": f.Deferred, "scanned": f.Scanned})
		fmt.Println(string(out))
		return
	}

	for _, r := range results {
		out, _ := json.Marshal(line{ID: r.ID, Neighbors: r.Neighbors, SumX: r.SumX, SumY: r.SumY})
		fmt.Println(string(out))
	}
	if *sample == "budget" {
		out, _ := json.Marshal(map[string]int{"computed": len(results), "deferred": f.Deferred})
		fmt.Println(string(out))
	}
}
