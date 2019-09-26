package limitedwriter

import (
	"bytes"
	"testing"
)

func Test_LimitedWriter(t *testing.T) {
	cap := 10
	buffer := new(bytes.Buffer)
	writer := NewLimitedWriter(buffer, cap)

	str := "123456789"

	num, err := writer.Write([]byte(str))
	if num != len(str) && err != nil {
		t.Fatalf("write normal data{%s} failed: err=%s n=%d", str, err, num)
	}
	t.Logf("write normal data{%s} succeeded. {num=%d err=%s}", str, num, err)

	num, err = writer.Write([]byte(str))
	if err != ErrOutofCapacity && num != 0 {
		t.Fatalf("write wrong data{%s} failed: err=%s n=%d", str, err, num)
	}
	t.Logf("write wrong data{%s} succeeded. {num=%d err=%s}", str, num, err)

	oneByte := "a"
	num, err = writer.Write([]byte(oneByte))
	if err != nil {
		t.Fatalf("write last byte{%s} failed: %s", oneByte, err)
	}
	t.Logf("write last byte{%s} succeeded. {err=%s}", str, err)

	num, err = writer.Write([]byte(str))
	if err != ErrOutofCapacity && num != 0 {
		t.Fatalf("write wrong data{%s} failed: err=%s n=%d", str, err, num)
	}
	t.Logf("write wrong data{%s} succeeded. {num=%d err=%s}", str, num, err)

	t.Logf("final buffer data: %s", buffer.String())
}
