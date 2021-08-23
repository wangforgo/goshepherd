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

type (
	shepherd struct {
		lock sync.Mutex
		sheep map[int]*exec.Cmd
	}
)

func NewShepherd() *shepherd {
	return &shepherd{
		sheep: make(map[int]*exec.Cmd),
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


type Rsp struct {
	StatusCode int
	Reason     string
}


type indexHandle struct {}

func (m *indexHandle) ServeHTTP (writer http.ResponseWriter, request *http.Request) {
	tmpl := template.New("index")
	indexFile, err := f.ReadFile("static/index.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	_,err = tmpl.Parse(string(indexFile))
	if err != nil {
		panic(err)
	}

	type Sheep struct {
		Name string
		Port int
		Path string
	}
	// todo range real sheep data
	tmpl.Execute(writer, Sheep{
				Name: "hello",
				Port: 123,
				Path: "uuuu",
	})
}

func main() {
	initGoToolPath()

	fmt.Println(getRandomPort())

	http.Handle("/static/", http.FileServer(http.FS(f)))
	http.Handle("/", &indexHandle{})
	http.Handle("/api", NewShepherd())

	log.Fatal(http.ListenAndServe(":7777", nil))
}



// initGoToolPath tries to find the path for pprof and trace.
func initGoToolPath() {
	toolPath := strings.TrimRight(strings.Replace(os.Getenv("GOROOT"),`\`,`/`,-1),"/") + "/pkg/tool/"
	myToolPath := toolPath+runtime.GOOS+"_"+runtime.GOARCH
	fmt.Println(myToolPath)

	filepath.Walk(myToolPath, func(path string, info fs.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(),"pprof") {
			pprofExePath = path
		} else if strings.HasPrefix(info.Name(),"trace") {
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


func bornSheep(cmd string, args ... string) (*exec.Cmd, string) {
	c := exec.Command(cmd, args...)
	ch := make(chan string, 1)
	go func() {
		output, err := c.CombinedOutput()
		if err != nil {
			ch <- string(output)
		}
	}()

	select {
	case output := <- ch:
		return nil, output
	//	we assume that all bad cases of opening go tools return within one second.
	case <- time.After(time.Second):
		return c, ""
	}
}

func purePath(path string) string {
	path = strings.Replace(path,`"`," ",-1)
	path = strings.Replace(path,"`"," ",-1)
	return strings.TrimSpace(path)
}

func (s *shepherd) add(v url.Values) string {
	path1 := purePath(v.Get("path1"))
	path2 := purePath(v.Get("path2"))

	var newSheep *exec.Cmd
	var err string

	randPort := getRandomPort()
	if randPort == 0 {
		return "port resource exhausted"
	}

	httpArgs := fmt.Sprintf("-http=127.0.0.1:%v", randPort)
	switch v.Get("tool") {
	case "0":
		newSheep, err = bornSheep(pprofExePath, httpArgs, path1)
	case "1":
		newSheep, err = bornSheep(traceExePath, httpArgs, path1)
	case "2":
		newSheep, err = bornSheep(pprofExePath, httpArgs, "-base", path1, path2)
	default:
		fmt.Printf("invalid tool type, got: %v\n",v.Get("tool"))
		return "invalid tool type"
	}

	if newSheep == nil {
		return err
	}

	s.lock.Lock()
	s.sheep[randPort] = newSheep
	s.lock.Unlock()

	return strconv.Itoa(randPort)
}


func (s *shepherd) rmv(v url.Values) string {
	port, err := strconv.Atoi(v.Get("port"))
	if err != nil || port == 0{
		return "invalid port"
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if c, ok := s.sheep[port]; ok && c!= nil {
		s.sheep[port].Process.Kill()
	}
	delete(s.sheep, port)
	return "ok"
}


func parseToolType(v url.Values) toolType {
	switch v.Get("tool") {
	case "0":
		return pprof
	case "1":
		return trace
	case "2":
		return pprof2
	default:
		return invalidTool
	}
}


func getRandomPort() int {
	l, err := net.Listen("tcp",":0")
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

