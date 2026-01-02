package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"fyne.io/fyne/v2/app"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits"

	"net/http"
	_ "net/http/pprof"
)

// Use `go build -tags matprofile` to profile mat leaks
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write heap (memory) profile to file")
var profserver = flag.Bool("profserver", false, "start profiling server; view at http://localhost:6060/debug/pprof/")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println("Error creating cpuprofile: ", err)
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}
	if *profserver {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
	a := app.NewWithID("com.github.sandalwoodbox.cleancredits")
	w := cleancredits.NewMainWindow(a)
	w.ShowAndRun()
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Println("Error creating memprofile: ", err)
		} else {
			pprof.WriteHeapProfile(f)
			f.Close()
		}
	}
}
