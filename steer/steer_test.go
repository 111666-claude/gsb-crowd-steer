package steer

import "testing"

func TestLoneAgentHasNoNeighbors(t *testing.T) {
	f := &Field{Radius: 2, MaxChecks: 1000, Workers: 2}
	got := f.Step([]Agent{{ID: "a", X: 0, Y: 0}})
	if len(got) != 1 || got[0].Neighbors != 0 {
		t.Fatalf("孤立代理不该有邻居：%+v", got)
	}
}

func TestCloseAgentIsCounted(t *testing.T) {
	f := &Field{Radius: 5, MaxChecks: 1000, Workers: 2}
	got := f.Step([]Agent{{ID: "a", X: 0, Y: 0}, {ID: "b", X: 1, Y: 0}})
	total := 0
	for _, r := range got {
		total += r.Neighbors
	}
	if total != 2 {
		t.Fatalf("两个代理应互相算作邻居，合计 2：%+v", got)
	}
}

func TestFarAgentIsNotCounted(t *testing.T) {
	f := &Field{Radius: 2, MaxChecks: 1000, Workers: 4}
	got := f.Step([]Agent{{ID: "a", X: 0, Y: 0}, {ID: "b", X: 100, Y: 100}})
	for _, r := range got {
		if r.Neighbors != 0 {
			t.Fatalf("远处的代理不该算邻居：%+v", got)
		}
	}
}

func TestSumPointsAtNeighbour(t *testing.T) {
	f := &Field{Radius: 5, MaxChecks: 1000, Workers: 1}
	got := f.Step([]Agent{{ID: "a", X: 0, Y: 0}, {ID: "b", X: 3, Y: 4}})
	for _, r := range got {
		if r.ID == "a" && (r.SumX != 3 || r.SumY != 4) {
			t.Fatalf("a 的分离向量应为 (3,4)：%+v", r)
		}
	}
}
