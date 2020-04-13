package supervisor

import (
	"bytes"
	"os/exec"
	"syscall"
)

// worker is just an abstraction over a unix command.
type worker struct {
	command *exec.Cmd
	stdout  string
	stderr  string
	running chan struct{}
	// count of restarts for a process
	restarts int
	// flag used to diffrentiate between crashes and termination by user
	restartToggle bool
	// process instance to be used during a restart
	process *Process
}

func newWorker(proc *Process, restarts int) *worker {
	// slices can be used in place of variadic arguments.
	// see: https://stackoverflow.com/questions/23723955/how-can-i-pass-a-slice-as-a-variadic-input
	cmd := exec.Command(proc.Executable, proc.Args...)
	cmd.Env = proc.Env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return &worker{
		command:       cmd,
		process:       proc,
		running:       make(chan struct{}, 1),
		restarts:      restarts,
		restartToggle: true,
	}
}

// starts the process waits for it to complete, if it crashes then signals
// a restart
func (w *worker) start(crashed chan *worker, completed chan struct{}) {
	var outbuf, errbuf bytes.Buffer
	w.command.Stdout = &outbuf
	w.command.Stderr = &errbuf

	err := w.command.Start()
	if err != nil {
		// return early since the command couldn't start it is possible
		// that the binary doesn't exist so we don't consider this a crash
		completed <- struct{}{}
		return
	}
	w.running <- struct{}{}

	// wait for the process to complete
	err = w.command.Wait()
	if err != nil {
		// signal a restart only in case of a crash
		// and not termination by user
		if w.restartToggle {
			crashed <- w
		}
		return
	}
	w.stdout = outbuf.String()
	w.stderr = errbuf.String()

	// signal successful completion
	completed <- struct{}{}
}

// terminates the process
func (w *worker) stop(completed chan struct{}) {
	<-w.running
	// See: https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly#29552044
	pgid, err := syscall.Getpgid(w.command.Process.Pid)
	if err != nil {
		return
	}
	syscall.Kill(-pgid, 15)
	// since it was stopped by the user and not crashed we
	// disable restarts for this worker
	w.restartToggle = false
	// in case of successfull termination we need to signal completion
	// so that the main Supervisor process can quit
	completed <- struct{}{}
}
