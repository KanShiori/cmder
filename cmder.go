package cmder

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"syscall"
	"time"
)

type State int

const (
	ErrCodeDefault       = 1001
	ErrCodeDetached      = 1002
	ErrCodeStartFailed   = 1003
	ErrCodeUnknownFailed = 1004

	OutputBufferMaxSize = 1024 * 1024

	Created  State = 1
	Running  State = 2
	Finished State = 3
)

var (
	UnRunningError = errors.New("cmder isn't running")
)

func IsInternalErrCode(code int) bool {
	switch code {
	case ErrCodeDefault, ErrCodeDetached, ErrCodeStartFailed, ErrCodeUnknownFailed:
		return true
	}

	return false
}

type Result struct {
	Code   int
	ErrMsg string
	Stdout string
	Stderr string
	Pid    int

	StartAt time.Time
	StopAt  time.Time
}

func (r Result) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// ExecuteIn 阻塞执行cmd命令, 工作目录位于dir下, 超时设定为 timeout
//
// cmd 超时会经过一个 detach 的过程, 所以真正命令返回的最大时间可能是 2 * timeout.
// 具体见 cmder.Execute 函数
func ExecuteIn(cmd string, timeout time.Duration, dir string) *Result {
	cmder := NewCmder("sh", "-c", cmd)
	cmder.SetDir(dir)
	return cmder.Execute(timeout)
}

// Execute 阻塞执行cmd命令, 超时设定为timeout
//
// cmd 超时会经过一个 detach 的过程, 所以真正命令返回的最大时间可能是2 * timeout.
// 具体见 cmder.Execute 函数
func Execute(cmd string, timeout time.Duration) *Result {
	cmder := NewCmder("sh", "-c", cmd)
	return cmder.Execute(timeout)
}

// TODO: 对于cmd status加锁
type Cmder struct {
	Name string
	Args []string

	cmd    *exec.Cmd
	status State

	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer

	result   Result
	doneChan chan struct{}
}

// Execute 执行Cmder的命令, 并阻塞等待命令返回或者超时
func (c *Cmder) Execute(timeout time.Duration) *Result {

	// 异步启动
	c.Start()

	// timeout时间内执行其命令的kill信号发送
	timer := time.AfterFunc(timeout, func() {
		c.Stop()
	})

	detachedTimer := time.NewTimer(2 * timeout)

	// 阻塞等待命令detach或者结束
	select {
	case <-detachedTimer.C:
		c.result.Code = int(ErrCodeDetached)
		c.result.ErrMsg = "timeout too long, detach cmd"
	case <-c.doneChan:
		timer.Stop()
		detachedTimer.Stop()
	}

	return &c.result
}

func (c *Cmder) Start() {
	// TODO: error?
	if c.status >= Running {
		return
	}

	go c.run()

	return
}

func (c *Cmder) SetDir(dir string) {
	c.cmd.Dir = dir
}

func (c *Cmder) RawCmd() *exec.Cmd {
	return c.cmd
}

func (c *Cmder) run() {
	// - final step: notify cmd finsish
	defer func() {
		close(c.doneChan)
		c.status = Finished
	}()

	// - start commond
	nowTime := time.Now()
	err := c.cmd.Start()
	if err != nil {
		c.result.Code = int(ErrCodeStartFailed)
		c.result.ErrMsg = err.Error()
		c.result.StartAt = nowTime
		c.result.StopAt = nowTime
		return
	}

	c.status = Running
	c.result.StartAt = nowTime
	c.result.Pid = c.cmd.Process.Pid

	// -  wait command Finished or be killed
	err = c.cmd.Wait()
	errcode := 0
	errmsg := "success"
	if err != nil {
		errmsg = err.Error()
		if exiterr, ok := err.(*exec.ExitError); ok {
			// ExitError表明非正常退出, 通过转换为 syscall.WaitStatus 获得 code
			// see https://stackoverflow.com/questions/10385551/get-exit-code-go
			if wstatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				errcode = wstatus.ExitStatus()
			}
		} else {
			errcode = int(ErrCodeUnknownFailed)
		}
	}

	// - set result
	c.result.Code = errcode
	c.result.ErrMsg = errmsg
	c.result.StopAt = time.Now()
	c.result.Stdout = c.stdoutBuffer.String()
	c.result.Stderr = c.stderrBuffer.String()
	return
}
