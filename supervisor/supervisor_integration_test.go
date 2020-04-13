package supervisor

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

var testProcesses = []*Process{
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
	{
		Executable: "non-existant-command",
		Args:       []string{"20"},
		Env:        nil,
	},
}

// all the processes are started, introduces a crash and a call to stop all
func TestSupervisor(t *testing.T) {
	sup := New(testProcesses, 2)
	done := sup.StartProcesses()

	// delay to make sure all the processes have started
	time.Sleep(1 * time.Second)

	for _, child := range sup.GetChildren() {
		proc := child.process
		if proc.Executable == "sleep" && proc.Args[0] == testProcesses[2].Args[0] {
			syscall.Kill(-child.command.Process.Pid, 15)
			break
		}
	}

	// delay to ensure the crashed process has restarted
	time.Sleep(1 * time.Second)

	<-done

	if sup.GetCrashedProcesses() != 1 {
		t.Errorf("Wrong number of successfull processes; got: %d, want: %d",
			sup.GetCrashedProcesses(), 1)
	}

	if sup.GetSuccessfullProcess() != len(testProcesses) {
		t.Errorf("Wrong number of successfull processes; got: %d, want: %d",
			sup.GetSuccessfullProcess(), len(testProcesses))
	}

	stdout := sup.ReadStdOut()
	stderr := sup.ReadStdErr()

	if len(stdout) != 2 {
		t.Errorf("Wrong number of stdout output, got: %d, want: %d", len(stdout), 2)
	}

	if len(stderr) != 1 {
		t.Errorf("Wrong number of stderr output, got: %d, want: %d", len(stderr), 1)
	}

	fmt.Println("starting all procs")
}
