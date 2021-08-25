# GoShepherd

## 1. Why GoShepherd?

`GoShepherd` is an efficient tool for managing go tool instances like pprof and trace. 

Before having GoShepherd, we may use the comand to open profile data files.

```go
$ go tool pprof -http=:7658 cpu.profile
```

There are three inconveniences：

- need to find an unused port by ourselves.
- it is hard to keep the relationship between port and opened data files.
- if we close the terminal, the opened pprof program doesn't exits, causeing a resource leak.

These problems become particularly obvious, when people are working with performance tuning and problem trouble-shooting.

Luckily, we have `GoShepherd` now, it addresses the problems well mentioned above.

## 2. How to use GoShepherd?

1. Download the souce code and build it with Go version (>=`1.16`).

2. Start the built execute `GoShepherd` bin.
3. Open projects in the home page of `GoShepherd`: `localhost:7777`.



TODO.