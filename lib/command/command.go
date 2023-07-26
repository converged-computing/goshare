package command

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Get the environment into a map
func GetEnvironment() map[string]string {
	vars := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		vars[pair[0]] = pair[1]
	}
	return vars
}

// Output is a basic structure for returning output, error, and exit code
type Output struct {
	Out      string
	Err      string
	ExitCode int
}

// A command holds the exec.Cmd and pointers to buffers, etc.
type CommandWrapper struct {
	Command *exec.Cmd
	Builder *strings.Builder
}

// RunInteractive performs a syscall (e.g., for a container shell)
func RunInteractive(command []string, env []string) error {

	environ := os.Environ()

	// If we have environment strings, add them
	if len(env) > 0 {
		environ = append(environ, env...)
	}
	// TODO add debug print here?
	return syscall.Exec(command[0], command, environ)
}

// Run a background process and return the command, response buffers, and any error
func RunDetachedCommand(command []string, env []string, workdir string) (CommandWrapper, error) {

	cmd, builder := exec.Command(command[0], command[1:]...), new(strings.Builder)
	cmd.Env = os.Environ()
	cmd.Stdout = builder

	// Set the working directory, if defined
	if workdir != "" {
		cmd.Dir = workdir
	}
	res := CommandWrapper{
		Command: cmd,
		Builder: builder,
	}
	err := cmd.Start()
	if err != nil {
		return res, err
	}
	return res, nil
}

// RunCommand runs one command and returs an error, output, and error
func RunCommand(command []string, env []string) (Output, error) {

	// Define the command!
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = os.Environ()

	// Prepare to write to output and error streams
	var outstream, errstream bytes.Buffer
	cmd.Stdout = &outstream
	cmd.Stderr = &errstream

	// If we have environment strings, add them
	if len(env) > 0 {
		cmd.Env = append(cmd.Env, env...)
	}

	// Run the command
	err := cmd.Run()

	// Prepare an output object to return
	output := Output{Out: outstream.String(), Err: errstream.String()}
	if err != nil {

		// Try to derive an exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			output.ExitCode = exitError.ExitCode()
		}
		return output, err
	}

	// Assume success without an error
	output.ExitCode = 0
	return output, nil
}
