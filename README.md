# crowd-steer

为一群代理计算邻域避让，按帧输出。只用 Go 标准库。

```
go test ./...
go run ./cmd/steer --sample=boundary
go run ./cmd/steer --sample=order
go run ./cmd/steer --sample=budget
go run ./cmd/steer --sample=work
```

## 口径

- **邻居**：两个代理的切比雪夫距离 `max(|dx|, |dy|)` 不大于 `Radius` 就算邻居，正好等于半径也算。
- **输出顺序**：结果一律按代理 ID 升序排列，与邻居数无关。
- **分离向量**：`sumx`、`sumy` 是各邻居相对本代理的坐标差之和。
- **检查预算**：一帧内累计检查的候选对数不超过 `MaxChecks`；超出后剩下的代理这一帧不输出，并计入 `Deferred`。
- **索引**：候选必须按网格分格取，只查本格与相邻格，不得两两配对。
- **幂等**：同一帧重复计算不重复计入检查数。

## 不变量

- 输出顺序只由代理 ID 决定，与 `Workers` 的并发调度无关。
- 被延后的代理不出现在结果里，`Deferred` 与输出条数之和等于代理总数。
- `Scanned` 不得随代理总数的平方放大。

## 输出契约

每个算完的代理一行：

```
{"id": <原样>, "neighbors": <邻居数>, "sumx": <x 方向和>, "sumy": <y 方向和>}
```

`--sample=budget` 末行打印 `{"computed": N, "deferred": N}`；
`--sample=work` 只打印 `{"agents": N, "deferred": N, "scanned": N}`。
