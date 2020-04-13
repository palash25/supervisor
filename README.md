# supervisor
Supervisor lets you start up a bunch of processes concurrently, stop them, restart them in case they fail or a killed by some other program and read their stdout and stderr. This is the submission for PerconaDB Backend Engineer Assignment.

## Design Decisions
- Restart logic is not a part of the public API. Since it was not clear from the task description I decided to keep the restart logic internal without exposing it to the user. In case of an error or a crash the process is restarted by the supervisor on its own. The user can only provide the restart limit through the public API.
- `Supervisor.StartProcesses` launches all the child processes in separate goroutines and two special goroutines.
    1. To capture a termination signal like `Ctrl+C` to kill the main process and all the children along with it.
    2. The second goroutine is sort of like a monitoring goroutine that listens on various channels for completion and crash messages from the workers and takes actions accordingly.
- The worker code communicates to the Supervisor using channels, this was done in accordance to Go's philosophy of "sharing by communicating not communicate by sharing"

## Potential Improvements
- Add a constant and exponential backoff strategy for retry, wasn't able to add this due to lack of time.
- Add error handling, tried to add an error chan but it seemed to hang up the process for some reason so had to abandon the idea.
- Due to lack of errors it was hard to write unit tests.
- Better test coverage. Its currently at 84.1%