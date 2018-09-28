package shutils

import (
	"os/exec"
	"fmt"
	"os"
	"bytes"
	"io"
	"github.com/pkg/errors"
	"syscall"
	"strings"
)

type ShCMD struct {
	stdout bytes.Buffer
	stderr bytes.Buffer

	X *exec.Cmd
	Args []string
}

func Cmd(name string, args ...interface{}) *ShCMD {
	var strArgs []string
	for _, arg := range args {
		strArgs = append(strArgs, fmt.Sprintf("%v", arg))
	}
	cmd := &ShCMD{X: exec.Command(name, strArgs...)}

	cmd.Args = append(cmd.Args, name)
	cmd.Args = append(cmd.Args, strArgs...)

	cmd.X.Stdout = &cmd.stdout
	cmd.X.Stderr = &cmd.stderr

	return cmd
}

func (cmd *ShCMD) SetENV(key, value string) (*ShCMD) {
	cmd.X.Env = append(cmd.X.Env, fmt.Sprintf("%v=%v", key, value))
	return cmd
}

func (cmd *ShCMD) Run() (*ShCMD) {
	cmd.X.Run()
	return cmd
}

func (cmd *ShCMD) PipeStdout(writer io.Writer) (*ShCMD) {
	cmd.X.Stdout = writer
	return cmd
}

func (cmd *ShCMD) PipeStderr(writer io.Writer) (*ShCMD) {
	cmd.X.Stderr = writer
	return cmd
}

func (cmd *ShCMD) Start() (*ShCMD) {
	cmd.X.Start()
	return cmd
}

func (cmd *ShCMD) Wait() (*ShCMD) {
	cmd.X.Wait()
	return cmd
}

func (cmd *ShCMD) Started() (bool) {
	return cmd.X.ProcessState != nil
}

func (cmd *ShCMD) Done() (bool) {
	return cmd.X.ProcessState != nil && cmd.X.ProcessState.Exited()
}

func (cmd *ShCMD) HasError() (bool) {
	return cmd.X.ProcessState != nil && cmd.X.ProcessState.Exited() && !cmd.X.ProcessState.Success()
}

func (cmd *ShCMD) State() (*os.ProcessState) {
	return cmd.X.ProcessState
}

func (cmd *ShCMD) StdoutStr() (string) {
	return strings.TrimSpace(cmd.stdout.String())
}

func (cmd *ShCMD) StderrStr() (string) {
	return strings.TrimSpace(cmd.stderr.String())
}

func (cmd *ShCMD) Value() (string, error) {
	return cmd.StdoutStr(), cmd.Err()
}

func (cmd *ShCMD) Err() (error) {
	if !cmd.HasError() {
		return nil
	}

	errorText := cmd.StderrStr()
	code := cmd.X.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()

	return errors.Errorf(
		"Command [%v] failed with status code (%d) STDERR:\n%v",
		strings.Join(cmd.Args, " "), code, errorText,
	)
}
