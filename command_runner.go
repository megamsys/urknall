package urknall

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/megamsys/urknall/cmd"
	"github.com/megamsys/urknall/target"
)

// commandRunner is used to execute commands in a build.
type commandRunner struct {
	build   *Build
	dir     string
	command cmd.Command

	taskName    string

	commandStarted time.Time
}

func (runner *commandRunner) run() error {
	runner.commandStarted = time.Now()

	checksum, e := commandChecksum(runner.command)
	if e != nil {
		return e
	}
	prefix := runner.dir + "/" + checksum

	if e = runner.writeScriptFile(prefix); e != nil {
		return e
	}

	errors := make(chan error)
	logs := runner.newLogWriter(prefix + ".log", errors)

	c, e := runner.build.prepareCommand("sh " + prefix + ".sh")
	if e != nil {
		return e
	}

	var wg sync.WaitGroup

	// Get pipes for stdout and stderr and forward messages to logs channel.
	stdout, e := c.StdoutPipe()
	if e != nil {
		return e
	}
	wg.Add(1)
	go runner.forwardStream(logs, "stdout", &wg, stdout)

	stderr, e := c.StderrPipe()
	if e != nil {
		return e
	}
	wg.Add(1)
	go runner.forwardStream(logs, "stderr", &wg, stderr)

	if sc, ok := runner.command.(cmd.StdinConsumer); ok {
		c.SetStdin(sc.Input())
		defer sc.Input().Close()
	}

	e = c.Run()
	wg.Wait()
	close(logs)

	// Get errors that might have occurred while handling the back-channel for the logs.
	for e = range errors {
		log.Printf("ERROR: %s", e.Error())
	}
	return e
}

func (runner *commandRunner) writeScriptFile(prefix string) (e error) {
	targetFile := prefix + ".sh"
	env := ""
	for _, e := range runner.build.Env {
		env += "export " + e + "\n"
	}
	rawCmd := fmt.Sprintf("cat <<\"EOSCRIPT\" > %s\n#!/bin/sh\nset -e\nset -x\n\n%s\n%s\nEOSCRIPT\n", targetFile, env, runner.command.Shell())
	c, e := runner.build.prepareInternalCommand(rawCmd)
	if e != nil {
		return e
	}

	return c.Run()
}

func logError(e error) {
	log.Printf("ERROR: %s", e.Error())
}

func (runner *commandRunner) forwardStream(logs chan <- string, stream string, wg *sync.WaitGroup, r io.Reader) {
	defer wg.Done()

	m := message("task.io", runner.build.hostname(), runner.taskName)
	m.Message = runner.command.Shell()
	if logger, ok := runner.command.(cmd.Logger); ok {
		m.Message = logger.Logging()
	}
	m.Stream = stream

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		m.Line = scanner.Text()
		if m.Line == "" {
			m.Line = " " // empty string would be printed differently therefore add some whitespace
		}
		m.TotalRuntime = time.Since(runner.commandStarted)
		m.Publish(stream)
		logs <- time.Now().UTC().Format(time.RFC3339Nano) + "\t" + stream + "\t" + scanner.Text()
	}
}

func (runner *commandRunner) newLogWriter(path string, errors chan <- error) chan <- string {
	logs := make(chan string)

	go func() {
		defer close(errors)

		cmd, err := runner.writeLogs(path, errors, logs)
		switch {
		case err != nil:
			errors <- err
		default:
			if err := cmd.Wait(); err != nil {
				errors <- err
			}
		}
	}()

	return logs
}

func (runner *commandRunner) writeLogs(path string, errors chan <- error, logs <-chan string) (target.ExecCommand, error) {
	// so ugly, but: sudo not required and "sh -c" adds some escaping issues with the variables. This is why Command is called directly.
	cmd, err := runner.build.Command("cat - > " + path)
	if err != nil {
		return nil, err
	}

	// Get pipe to stdin of the execute command.
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	defer in.Close()

	// Run command, writing everything coming from stdin to a file.
	if err := cmd.Start(); err != nil {
		in.Close()
		return nil, err
	}

	// Send all messages from logs to the stdin of the new session.
	for log := range logs {
		if _, err = io.WriteString(in, log + "\n"); err != nil {
			errors <- err
		}
	}
	return cmd, err

}
