package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type toolType int

const (
	pprof toolType = iota
	trace
	pprof2
	invalidTool
)

//go:embed static/*
var f embed.FS
var pprofExePath, traceExePath string
var shepherdInst = &shepherd{
	sheep: make(map[int]*Sheep),
}

type (
	shepherd struct {
		lock  sync.Mutex
		sheep map[int]*Sheep
	}

	Sheep struct {
		inst  *exec.Cmd // command instance of go tools
		name  string    // project name
		path1 string    // path of the first file
		path2 string    // path of the second file, only needed when comparing two files
		port  int       // assigned port for the tool
	}
)

func (s *shepherd) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()

	// check op
	var rsp string
	switch query.Get("op") {
	case "add":
		rsp = s.add(query)
	case "rmv":
		rsp = s.rmv(query)
	default:
		rsp = "op not support"
	}

	writer.Write([]byte(rsp))
}

type indexHandle struct{}

func (m *indexHandle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	tmpl := template.New("index")
	indexFile, err := f.ReadFile("static/index.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = tmpl.Parse(string(indexFile))
	if err != nil {
		panic(err)
	}

	type SheepForHtml struct {
		Name  string
		Port  int
		Path1 string
		Path2 string
	}

	liveSheep := shepherdInst.allLiveSheep()
	allSheepHtml := make([]SheepForHtml, len(liveSheep))
	for i, s := range liveSheep {
		allSheepHtml[i] = SheepForHtml{
			Name:  s.name,
			Port:  s.port,
			Path1: s.path1,
			Path2: s.path2,
		}
	}

	tmpl.Execute(writer, allSheepHtml)
}

func main() {
	initGoToolPath()

	http.Handle("/static/", http.FileServer(http.FS(f)))
	http.Handle("/", &indexHandle{})
	http.Handle("/api", shepherdInst)

	log.Fatal(http.ListenAndServe(":7777", nil))
}

// initGoToolPath tries to find the path for pprof and trace.
func initGoToolPath() {
	toolPath := strings.TrimRight(strings.Replace(os.Getenv("GOROOT"), `\`, `/`, -1), "/") + "/pkg/tool/"
	myToolPath := toolPath + runtime.GOOS + "_" + runtime.GOARCH
	fmt.Println(myToolPath)

	filepath.Walk(myToolPath, func(path string, info fs.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "pprof") {
			pprofExePath = path
		} else if strings.HasPrefix(info.Name(), "trace") {
			traceExePath = path
		}
		return nil
	})

	if pprofExePath == "" {
		fmt.Printf("cannot find executable tool: pprof!\n")
		os.Exit(-1)
	}

	if traceExePath == "" {
		fmt.Printf("cannot find executable tool: trace!\n")
		os.Exit(-1)
	}
}

func runCmd(cmd string, args ...string) (*exec.Cmd, string) {
	c := exec.Command(cmd, args...)
	ch := make(chan string, 1)
	go func() {
		output, err := c.CombinedOutput()
		if err != nil {
			ch <- string(output)
		}
	}()

	select {
	case output := <-ch:
		return nil, output
	//	we assume that all bad cases of opening go tools return within one second.
	case <-time.After(time.Second):
		return c, ""
	}
}

func purePath(path string) string {
	path = strings.Replace(path, `"`, " ", -1)
	path = strings.Replace(path, "`", " ", -1)
	return strings.TrimSpace(path)
}

func (s *shepherd) allLiveSheep() []*Sheep {
	allSheep := make([]*Sheep, 0, 0)
	s.lock.Lock()
	for _, v := range s.sheep {
		allSheep = append(allSheep, v)
	}
	s.lock.Unlock()
	return allSheep
}

func (s *shepherd) add(v url.Values) string {
	path1 := purePath(v.Get("path1"))
	path2 := purePath(v.Get("path2"))

	var cmdInst *exec.Cmd
	var err string

	randPort := getRandomPort()
	if randPort == 0 {
		return "port resource exhausted"
	}

	httpArgs := fmt.Sprintf("-http=127.0.0.1:%v", randPort)
	switch v.Get("tool") {
	case "0":
		cmdInst, err = runCmd(pprofExePath, httpArgs, path1)
	case "1":
		cmdInst, err = runCmd(traceExePath, httpArgs, path1)
	case "2":
		cmdInst, err = runCmd(pprofExePath, httpArgs, "-base", path1, path2)
	default:
		fmt.Printf("invalid tool type, got: %v\n", v.Get("tool"))
		return "invalid tool type"
	}

	if cmdInst == nil {
		return err
	}

	newSheep := &Sheep{
		inst:  cmdInst,
		name:  v.Get("name"),
		path1: path1,
		path2: path2,
		port:  randPort,
	}

	s.lock.Lock()
	s.sheep[randPort] = newSheep
	s.lock.Unlock()

	return strconv.Itoa(randPort)
}

func getRandomPort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Printf("failed to get random port, err: %v\n", err)
		return 0
	}
	defer l.Close()

	a, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0
	}

	return a.Port
}

func (s *shepherd) rmv(v url.Values) string {
	port, err := strconv.Atoi(v.Get("port"))
	if err != nil || port == 0 {
		return "invalid port"
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if c, ok := s.sheep[port]; ok && c != nil {
		s.sheep[port].inst.Process.Kill()
	}
	delete(s.sheep, port)
	return "ok"
}
