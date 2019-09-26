package limitedwriter

import (
	"errors"
	"io"
)

var ErrOutofCapacity = errors.New("LimitedWriter: out of capacity ")

// LimitedWriter 从W中读取数据，但限制可以读取的数据的量为最多N字节，每次调用Write方法都会更新N以标记剩余可以写入的字节数
//
// LimitedWriter 只限制了从LimitedWriter写入的数据量, 不会限制底层真正的大小
type LimitedWriter struct {
	W io.Writer
	N int
}

func NewLimitedWriter(w io.Writer, n int) *LimitedWriter {
	return &LimitedWriter{
		W: w,
		N: n,
	}
}

// 实现了io.Writer, 如果len(p)大于剩余可写字节数, 会拒绝写入并返回ErrOutofCapacity
func (lb *LimitedWriter) Write(p []byte) (n int, err error) {
	if len(p) > lb.N {
		return 0, ErrOutofCapacity
	}

	n, err = lb.W.Write(p)
	lb.N = lb.N - n
	return
}

func (lb *LimitedWriter) Rest() int {
	return lb.N
}
