package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed static/*
var f embed.FS
var pprofExePath, traceExePath string

func main() {
	welcome()
	initGoToolPath()

	http.Handle("/", &indexHandle{})
	http.Handle("/static/", http.FileServer(http.FS(f)))
	http.Handle("/api", shepherdInst)

	ch := make(chan int)
	go func() {
		select {
		case <- ch:
		case <- time.After(time.Millisecond*300):
			startHomePage() // only start home page when everything is ready.
		}
	}()

	if err := http.ListenAndServe(":7777", nil); err != nil {
		ch <- 1
		log.Fatal(err)
	}
}

type indexHandle struct{}

func (h *indexHandle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
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

// initGoToolPath find the path for pprof and trace.
func initGoToolPath() {
	toolPath := strings.TrimRight(strings.Replace(os.Getenv("GOROOT"), `\`, `/`, -1), "/") + "/pkg/tool/"
	goToolPath := toolPath + runtime.GOOS + "_" + runtime.GOARCH

	filepath.Walk(goToolPath, func(path string, info fs.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "pprof") {
			pprofExePath = path
		} else if strings.HasPrefix(info.Name(), "trace") {
			traceExePath = path
		}
		return nil
	})

	if pprofExePath == "" {
		log.Fatalf("pprof not found in %v\n", goToolPath)
	}

	if traceExePath == "" {
		log.Fatalf("trace not found in %v\n", goToolPath)
	}
}

func welcome() {
	fmt.Println("\n  _____        ____   __                __                 __\n / ___/ ___   / __/  / /  ___    ___   / /  ___   ____ ___/ /\n/ (_ / / _ \\ _\\ \\   / _ \\/ -_)  / _ \\ / _ \\/ -_) / __// _  / \n\\___/  \\___//___/  /_//_/\\__/  / .__//_//_/\\__/ /_/   \\_,_/  \n                              /_/                            ")
	fmt.Println("Welcome to GoShepherd!")
}

func startHomePage() {
	homepage := "http://127.0.0.1:7777"
	commands := map[string]string{
		"windows": "explorer",
		"darwin":  "open",
		"linux":   "xdg-open",
	}

	explorer, ok := commands[runtime.GOOS]
	if !ok {
		fmt.Println("Please visit home page manually: ", homepage)
		return
	}
	fmt.Println("Opening home page: ", homepage)

	exec.Command(explorer, homepage).Run()
}