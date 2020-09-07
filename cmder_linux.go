// +build  !windows

package cmder

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"

	"github.com/YshShiori/cmder/limitedwriter"
)

func NewCmder(name string, args ...string) *Cmder {
	cmder := &Cmder{
		Name:         name,
		Args:         args,
		cmd:          nil,
		status:       Create,
		stdoutBuffer: new(bytes.Buffer),
		stderrBuffer: new(bytes.Buffer),
		doneChan:     make(chan struct{}),
		result: Result{
			Code:    int(ErrorDefaultFailed),
			ErrMsg:  "default msg",
			Stdout:  "",
			Stderr:  "",
			Pid:     -1,
			StartTs: -1,
			StopTs:  -1,
		},
	}

	// set cmd
	cmder.cmd = exec.Command(cmder.Name, cmder.Args...)
	cmder.cmd.Stdout = limitedwriter.NewLimitedWriter(cmder.stdoutBuffer, OutputBufferMaxSize)
	cmder.cmd.Stderr = limitedwriter.NewLimitedWriter(cmder.stderrBuffer, OutputBufferMaxSize)

	// cmd进程创建一个新的 process group, 使得在stop时能够通过pgid kill cmd以及子进程.
	// 如果不kill子进程, 会导致Wait还是等待子进程的结束.
	cmder.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// 设置stdout每行打印长度
	cmder.cmd.Env = os.Environ()
	cmder.cmd.Env = append(cmder.cmd.Env, "COLUMNS=512")

	return cmder
}

func (c *Cmder) Stop() error {
	if c.status == Create || c.status == Finish {
		return UnRunningError
	}

	// kill process group
	pgid, err := syscall.Getpgid(c.cmd.Process.Pid)
	if err == nil {
		return syscall.Kill(-pgid, syscall.SIGTERM)
	}
	return err
}
