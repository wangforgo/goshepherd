# GoShepherd

## 1. Why GoShepherd?

`GoShepherd` is an efficient tool for managing go tool instances, i.e., pprof and trace. 

Before having GoShepherd, we may use the comand to open profile data files.

```bash
$ go tool pprof -http=:7658 cpu.profile
```

There are three inconveniences：

- we need to find an idle port by ourselves.
- it's hard to record the relationship between port and already opened data files.
- if we shut down the terminal, the opened go tool instance doesn't exits, causeing resource leaking.

These inconveniences become particularly obvious, when people are working with performance tuning and problem trouble-shooting.

Luckily, we have `GoShepherd` now, it addresses the problems mentioned above very vell.



## 2. How to use GoShepherd?

Prerequisites to check:

- require Go installed already. GoShepherd will check whether there are executable go tools below the `GOPATH` directory.
- tools for go tools are already installed, Graphviz is required for pprof.

To check the conditions above, try to open the cpu.profile provided by command. If the cpu.profile can be shown in the explorer, then all the prerequisites  are meet.

Then we can start to use GoShepherd:

1. Download the souce code and build it with Go version (>=`1.16`). For windows users, you may only download the built `goshepherd.exe`.

2. Run the built execute `GoShepherd` bin.
3. Open projects in the home page of `GoShepherd`: `127.0.0.1:7777`. Choose the tool type, name your project and input the path of the opening file, then click the `OPEN` button.



## 3. Tips

TODO.

