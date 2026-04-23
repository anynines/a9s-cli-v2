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

type Command struct {
	name           string
	args           []string
	env            []string
	stdIn          []byte
	commandMode    CommandMode
	ctx            context.Context
	interactive    bool
	suppressOutput bool
}

func NewCommand(name string, args ...string) *Command {
	return &Command{
		commandMode: CommandModeWithPrompt,
		name:        name,
		args:        args,
	}
}

func (c *Command) Ctx(ctx context.Context) *Command {
	c.ctx = ctx
	return c
}

func (c *Command) WithPrompt() *Command {
	c.commandMode = CommandModeWithPrompt
	return c
}

func (c *Command) NoPrompt() *Command {
	c.commandMode = CommandModeNoPrompt
	return c
}

func (c *Command) Quiet() *Command {
	c.commandMode = CommandModeQuiet
	return c
}

func (c *Command) GetName() string {
	return c.name
}

func (c *Command) GetArgs() []string {
	return c.args
}

// Interactive connects the command's stdin/stdout/stderr directly to the
// terminal. Use this for programs (like pagers) that need to take over the TTY.
func (c *Command) Interactive() *Command {
	c.interactive = true
	return c
}

func (c *Command) SuppressOutput() *Command {
	c.suppressOutput = true
	return c
}

func (c *Command) Env(env string, envs ...string) *Command {
	c.env = append([]string{env}, envs...)
	return c
}

func (c *Command) Stdin(stdIn []byte) *Command {
	c.stdIn = stdIn
	return c
}

func (c *Command) Run() ([]byte, error) {
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

	if (Verbose || err != nil) && !c.suppressOutput {
		fmt.Println(string(output))
	}

	return output, err
}

func printCommandString(command *exec.Cmd, c *Command) {

	if c.commandMode == CommandModeQuiet || !ShowCommands {
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
		if Verbose {
			PrintSmallCommand(commandString)
		}
		return
	}

	err := fmt.Errorf("No Command Mode chosen for command %s", commandString)
	ExitDueToFatalError(err, "")
}
