package cmder

import (
	"testing"
	"time"
)

func Test_Cmder(t *testing.T) {
	cmd := "ls"
	dir := ""
	timeout := 10 * time.Second

	t.Logf("try to exec cmd %s in dir %s with timeout %d ...\n", cmd, dir, timeout)
	Result := ExecuteIn(cmd, timeout, dir)
	t.Logf("result of exec:\n")
	t.Logf("  pid:\n    %d\n", Result.Pid)
	t.Logf("  startTs:\n    %v\n", Result.StartAt)
	t.Logf("  stopTs:\n    %v\n", Result.StopAt)
	t.Logf("  code:\n    %d\n", Result.Code)
	t.Logf("  errmsg:\n    %s\n", Result.ErrMsg)
	t.Logf("  stdout:\n    %s\n", Result.Stdout)
	t.Logf("  stderr:\n    %s\n", Result.Stderr)
}
