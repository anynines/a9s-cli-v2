package makeup

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CommandMode string

var (
	CommandModeDefault    CommandMode = "Default"
	CommandModeWithPrompt CommandMode = "WithPrompt"
	CommandModeNoPrompt   CommandMode = "NoPrompt"
	CommandModeQuiet      CommandMode = "Quiet"
)

var (
	ExecCommand        = exec.Command
	ExecCommandContext = exec.CommandContext
)

type command struct {
	name        string
	args        []string
	env         []string
	stdIn       []byte
	commandMode CommandMode
	ctx         context.Context
	interactive bool
}

func Command(name string, args ...string) command {
	return command{
		commandMode: CommandModeWithPrompt,
		name:        name,
		args:        args,
	}
}

func (c command) Ctx(ctx context.Context) command {
	result := c
	result.ctx = ctx
	return result
}

func (c command) WithPrompt() command {
	result := c
	result.commandMode = CommandModeWithPrompt
	return result
}

func (c command) NoPrompt() command {
	result := c
	result.commandMode = CommandModeNoPrompt
	return result
}

func (c command) Quiet() command {
	result := c
	result.commandMode = CommandModeQuiet
	return result
}

// Interactive connects the command's stdin/stdout/stderr directly to the
// terminal. Use this for programs (like pagers) that need to take over the TTY.
func (c command) Interactive() command {
	result := c
	result.interactive = true
	return result
}

func (c command) Env(env string, envs ...string) command {
	result := c
	result.env = append([]string{env}, envs...)
	return result
}

func (c command) Stdin(stdIn []byte) command {
	result := c
	result.stdIn = stdIn
	return result
}

func (c command) Run() ([]byte, error) {
	var command *exec.Cmd
	if c.ctx != nil {
		command = ExecCommandContext(c.ctx, c.name, c.args...)
	} else {
		command = ExecCommand(c.name, c.args...)
	}

	if c.stdIn != nil {
		command.Stdin = bytes.NewReader(c.stdIn)
	}

	if c.env != nil {
		command.Env = append(os.Environ(), c.env...)
	}

	printCommandString(command, c)

	// Interactive mode: connect directly to the terminal so programs like
	// pagers can take over the TTY. CombinedOutput() would capture the output
	// and prevent the program from rendering to the screen.
	if c.interactive {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if c.stdIn != nil {
			command.Stdin = bytes.NewReader(c.stdIn)
		} else {
			command.Stdin = os.Stdin
		}
		return nil, command.Run()
	}

	output, err := command.CombinedOutput()

	if Verbose || err != nil {
		fmt.Println(string(output))
	}

	return output, err
}

func printCommandString(command *exec.Cmd, c command) {

	if c.commandMode == CommandModeQuiet {
		return
	}

	quotedArgs := []string{}
	for i := range command.Args {
		arg := command.Args[i]
		if strings.Contains(arg, " ") || strings.Contains(arg, "$") {
			arg = `"` + arg + `"`
		}
		quotedArgs = append(quotedArgs, arg)
	}
	commandString := strings.Join(quotedArgs, " ")

	if c.commandMode == CommandModeWithPrompt {
		if UnattendedMode {
			PrintSmallCommand(commandString)
			return
		}
		PrintCommandBox(commandString)
		if c.stdIn != nil {
			println("Standard Input:\n" + string(c.stdIn))
		}
		WaitForUser()
		return
	}

	if c.commandMode == CommandModeNoPrompt {
		PrintSmallCommand(commandString)
		return
	}

	err := fmt.Errorf("No Command Mode chosen for command %s", commandString)
	ExitDueToFatalError(err, "")
}
