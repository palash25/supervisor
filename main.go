package main

import (
	"syscall"
	"time"

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
			Args:       []string{"10"},
			Env:        nil,
		},
	}

	sup := supervisor.New(processes, 2)
	done := sup.StartProcesses()

	// short delay to ensure that all the processes have started
	time.Sleep(2 * time.Second)

	pids := sup.GetPIDs()
	// cause a crash to simulate a restart
	_ = syscall.Kill(-pids["sleep"], 15) // get a print statement saying "restarted"

	// short delay to let the crashed process restart before stopping
	time.Sleep(1 * time.Second)
	// stop all running processes
	sup.StopProcesses()

	<-done
	//fmt.Println(sup.ReadStdErr())
	//fmt.Println(sup.ReadStdOut())
}
