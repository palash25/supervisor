package main

import (
	"fmt"

	"github.com/palash25/supervisor/supervisor"
)

func main() {
	processes := []*supervisor.Process{
		{
			Executable: "ls",
			Args:       []string{"-a", "."},
			Env:        nil,
		},
		{
			Executable: "curl",
			Args:       []string{"golang.org"},
			Env:        nil,
		},
		{
			Executable: "sleep",
			Args:       []string{"5"},
			Env:        nil,
		},
	}

	sup := supervisor.New(processes, 2)
	done := sup.StartProcesses()

	//time.Sleep(5 * time.Second)
	//fmt.Println(sup.ReadProcessStdErr(processes[4]))
	//fmt.Println(sup.ReadProcessStdOut(processes[4]))

	//time.Sleep(5 * time.Second)
	//sup.StopProcesses()

	<-done
	fmt.Println("starting all procs")
}
