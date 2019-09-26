package simplecmder

import (
	"bytes"
	"errors"
	"os/exec"
	"syscall"
	"time"
)

const (
	// ErrorDefaultFailed 错误默认值(不会抛出)
	ErrorDefaultFailed = 1001
	// ErrorCmdDetached cmd执行超过2*timeout, detach cmd
	ErrorCmdDetached = 1002
	// ErrorCmdStartFailed cmd启动失败
	ErrorCmdStartFailed = 1003
	// ErrorCmdStartFailed cmd执行未知错误
	ErrorCmdUnknownFailed = 1004

	// OutputBufferMaxSize 输出结果的最大大小
	OutputBufferMaxSize = 1024 * 1024
)

var (
	// UnRunningError 当Stop不在运行的cmder时返回
	UnRunningError = errors.New("cmder isn't running")
)

const (
	Create  = 1
	Running = 2
	Finish  = 3
)

func ExecuteCmd(cmd string, dir string, timeout int) *Result {
	cmder := NewCmder("sh", "-c", cmd)
	if dir != "" {
		cmder.SetDir(dir)
	}
	return cmder.Execute(timeout)
}

// TODO: 对于cmd status加锁
type Cmder struct {
	Name string
	Args []string

	cmd    *exec.Cmd
	status int

	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer
	result       Result
	doneChan     chan struct{}
}

// Result 是一个Cmder执行后的结果
// 其中:
//	Code Cmder返回的错误码 或 cmd返回的错误码
// 	ErrMsg Cmder返回的错误信息
//	Stdout 执行cmd的stdout
//	Stderr 执行cmd的stderr
//	Pid 执行cmd的Pid
//	StartTs cmd执行启动时间戳
//	StopTs cmd执行结束时间戳
type Result struct {
	Code    int
	ErrMsg  string
	Stdout  string
	Stderr  string
	Pid     int
	StartTs int64
	StopTs  int64
}

// Execute 执行Cmder的命令
func (c *Cmder) Execute(timeout int) *Result {

	// - start
	c.Start()

	// - stop 计时器
	timer := time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		c.Stop()
	})

	// - detach 计时器
	detachedTimer := time.NewTimer(time.Duration(2*timeout) * time.Second)

	// - 等待cmd结束 / 超过detach计时
	select {
	case <-detachedTimer.C:
		c.result.Code = int(ErrorCmdDetached)
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

func (c *Cmder) run() {
	// --- final step: notify cmd finsish
	defer func() {
		close(c.doneChan)
		c.status = Finish
	}()

	// --- start commond
	nowTs := time.Now().Unix()
	err := c.cmd.Start()
	if err != nil {
		c.result.Code = int(ErrorCmdStartFailed)
		c.result.ErrMsg = err.Error()
		c.result.StartTs = nowTs
		c.result.StopTs = nowTs
		return
	}

	c.status = Running
	c.result.StartTs = nowTs
	c.result.Pid = c.cmd.Process.Pid

	// --- wait command finish or be killed
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
			errcode = int(ErrorCmdUnknownFailed)
		}
	}

	// --- set result
	c.result.Code = errcode
	c.result.ErrMsg = errmsg
	c.result.StopTs = time.Now().Unix()
	c.result.Stdout = c.stdoutBuffer.String()
	c.result.Stderr = c.stderrBuffer.String()
	return
}
