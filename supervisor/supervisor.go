package supervisor

import (
	"fmt"
	"os"
	"os/signal"
)

// Supervisor maintains the state of all processes and manages them.
type Supervisor struct {
	// chan for trapping a termination/interrupt signal
	trap chan os.Signal
	// chan for a successfull completion signal
	completed chan struct{}
	// chan for a crash signal
	crashed chan *worker
	// chan to signal when all the processes have either completed
	// or been killed
	quit chan struct{}
	// a state of all child processes (running or dead)
	children []*worker
	// count of successfully completed or killed processes
	successfull int
	// keep a count crashes
	crashes int
	// total number of processes provided by the user to run initially
	total int
	// maximum number of restarts allowed for a process
	maxRestarts int
}

// New creates a new instance of the Supervisor. It accepts a list of
// processes to run along with the maximum number of restarts
func New(processes []*Process, maxRestarts int) *Supervisor {
	var childProcesses []*worker
	for _, proc := range processes {
		// pass `0` as the restart when initializing processes first time
		childProcesses = append(childProcesses, newWorker(proc, 0))
	}
	numProcs := len(processes)

	return &Supervisor{
		trap:        make(chan os.Signal, 1),
		completed:   make(chan struct{}, numProcs),
		crashed:     make(chan *worker, numProcs),
		children:    childProcesses,
		successfull: 0,
		crashes:     0,
		quit:        make(chan struct{}, 1),
		total:       numProcs,
		maxRestarts: maxRestarts,
	}
}

// StartProcesses starts all the child processes
func (s *Supervisor) StartProcesses() chan struct{} {
	done := make(chan struct{})
	// capture the signal to terminate the main process and its children
	signal.Notify(s.trap, os.Interrupt, os.Kill)
	go func() {
		<-s.trap
		fmt.Println("received kill signal")
		// kill all the child processes before killing the parent
		s.StopProcesses()
		fmt.Println("killed all")
		os.Exit(1)
	}()

	// start all the child processes in background
	for _, child := range s.children {
		childWorker := child
		go childWorker.start(s.crashed, s.completed)
	}

	// monitoring goroutine to listen for crashes/completions
	// of all running processes
	go func() {
		for {
			select {
			case <-s.completed:
				s.successfull++
				if s.successfull == s.total {
					s.quit <- struct{}{}
				}
			case worker := <-s.crashed:
				s.restartProcess(worker)
			case <-s.quit:
				fmt.Println("quit")
				done <- struct{}{}
				close(s.completed)
				close(s.crashed)
				return
			}
		}
	}()
	return done
}

// StopProcesses can be used to stop all the child processes.
func (s *Supervisor) StopProcesses() {
	fmt.Println(len(s.children))
	for _, child := range s.children {
		child.stop(s.completed)
	}
}

// ReadStdOut can be used to read the output of all the worker processes
func (s *Supervisor) ReadStdOut() map[string]string {
	result := make(map[string]string)
	for _, child := range s.children {
		if child.stdout != "" {
			if _, ok := result[child.process.Executable]; !ok {
				result[child.process.Executable] = child.stdout
				continue
			}
			result[child.process.Executable] += child.stdout
		}
	}
	return result
}

// ReadStdErr can be used to read the error output of all the worker processes
func (s *Supervisor) ReadStdErr() map[string]string {
	result := make(map[string]string)
	for _, child := range s.children {
		if child.stderr != "" {
			if _, ok := result[child.process.Executable]; !ok {
				result[child.process.Executable] = child.stderr
				continue
			}
			result[child.process.Executable] += child.stderr
		}
	}
	return result
}

// restart worker in case of a crash
func (s *Supervisor) restartProcess(proc *worker) {
	fmt.Println("restarting ", proc.command.Path)
	// record the crash
	s.crashes++
	for index, child := range s.children {
		if proc == child {
			fmt.Println(proc.restarts)
			if proc.restarts > s.maxRestarts {
				// maximum threshold for restarts reached, we therefore
				// consider this process terminated.
				s.completed <- struct{}{}
				return
			}
			// replace the old worker with the new one
			s.children[index] = newWorker(proc.process, proc.restarts+1)
			go s.children[index].start(s.crashed, s.completed)
			return
		}
	}
}

func (s *Supervisor) GetSuccessfullProcess() int {
	return s.successfull
}

func (s *Supervisor) GetCrashedProcesses() int {
	return s.crashes
}

func (s *Supervisor) GetChildren() []*worker {
	return s.children
}
