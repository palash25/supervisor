package supervisor

// Process is a type that the user can use to define the
// process and its args and env vars to run
type Process struct {
	Executable string
	Args       []string
	Env        []string
}
