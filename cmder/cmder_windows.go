// +build  windows

package cmder

import (
	"bytes"
	"os/exec"
	"strconv"

	"github.com/YshShiori/simple_cmder/limitedwriter"
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

	// windows 不支持
	// cmder.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmder
}

func (c *Cmder) Stop() error {
	// TODO: 解决Stop可能被多次调用的问题
	if c.status == Create || c.status == Finish {
		return UnRunningError
	}

	// kill process group
	// TODO: windwos目前只是简单的尝试杀死
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(c.cmd.Process.Pid))
	kill.Run()
	return nil
}
