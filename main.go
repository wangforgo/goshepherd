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
var shepherdInst = newShepherd()

type (
	shepherd struct {
		lock  sync.Mutex
		head *Sheep
		tail *Sheep
	}

	Sheep struct {
		next *Sheep
		inst  *exec.Cmd // command instance of go tools
		name  string    // project name
		path1 string    // path of the first file
		path2 string    // path of the second file, only needed when comparing two files
		port  int       // assigned unique port for the tool
	}
)

func newShepherd() *shepherd {
	dummy := &Sheep{}
	return &shepherd{
		head: dummy,
		tail: dummy,
	}
}

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

	liveSheep := shepherdInst.dumpSheep()
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

	s.addSheep(&Sheep{
		inst:  cmdInst,
		name:  v.Get("name"),
		path1: path1,
		path2: path2,
		port:  randPort,
	})

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
	s.rmvSheep(port)
	return "ok"
}

// addSheep should add new sheep with global unique port
func (s *shepherd) addSheep(sheep *Sheep) {
	s.lock.Lock()
	s.lock.Unlock()
	s.tail.next = sheep
	s.tail = sheep

}

func (s *shepherd) rmvSheep(port int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	prev := s.head
	next := s.head.next
	for next != nil {
		if next.port != port {
			prev = next
			next = next.next
			continue
		}
		prev.next = next.next
		if next == s.tail {
			s.tail = prev
		}
		break
	}
}

func (s *shepherd) dumpSheep() []*Sheep {
	allSheep := make([]*Sheep, 0, 0)
	s.lock.Lock()
	defer s.lock.Unlock()
	next := s.head.next
	for next != nil {
		allSheep = append(allSheep, next)
		next = next.next
	}
	return allSheep
}
