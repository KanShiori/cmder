package cmder

import (
	"testing"
)

func Test_Cmder(t *testing.T) {
	cmd := "ls"
	dir := ""
	timeout := 10

	t.Logf("try to exec cmd %s in dir %s with timeout %d ...\n", cmd, dir, timeout)
	Result := ExecuteCmd(cmd, dir, timeout)
	t.Logf("result of exec:\n")
	t.Logf("  pid:\n    %d\n", Result.Pid)
	t.Logf("  startTs:\n    %d\n", Result.StartTs)
	t.Logf("  stopTs:\n    %d\n", Result.StopTs)
	t.Logf("  code:\n    %d\n", Result.Code)
	t.Logf("  errmsg:\n    %s\n", Result.ErrMsg)
	t.Logf("  stdout:\n    %s\n", Result.Stdout)
	t.Logf("  stderr:\n    %s\n", Result.Stderr)
}
