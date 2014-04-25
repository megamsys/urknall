package local

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/dynport/urknall"
)

type Host struct {
	cachedUser string
}

func (c *Host) String() string {
	return "LOCAL"
}

func (c *Host) User() string {
	if c.cachedUser == "" {
		out := &bytes.Buffer{}
		err := &bytes.Buffer{}
		cmd := exec.Command("whoami")
		cmd.Stdout = out
		cmd.Stderr = err
		e := cmd.Run()
		if e != nil {
			fmt.Printf("error reading login name: err=%q out=%q e=%q", err.String(), out.String, e)
			os.Exit(1)
		}
		c.cachedUser = out.String()
	}
	return c.cachedUser
}

func (c *Host) Command(cmd string) (urknall.Command, error) {
	return &Command{
		command: exec.Command("bash", "-c", cmd),
	}, nil
}

type Command struct {
	command *exec.Cmd
}

func (c *Command) StdoutPipe() (io.Reader, error) {
	return c.command.StdoutPipe()
}

func (c *Command) StderrPipe() (io.Reader, error) {
	return c.command.StderrPipe()
}

func (c *Command) StdinPipe() (io.Writer, error) {
	return c.command.StdinPipe()
}

func (c *Command) SetStdout(w io.Writer) {
	c.command.Stdout = w
}

func (c *Command) SetStderr(w io.Writer) {
	c.command.Stderr = w
}

func (c *Command) SetStdin(r io.Reader) {
	c.command.Stdin = r
}

func (c *Command) Wait() error {
	return c.command.Wait()
}

func (c *Command) Start() error {
	return c.command.Start()
}

func (c *Command) Run() error {
	return c.command.Run()
}