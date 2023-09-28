package socks5

import (
	"errors"
	"github.com/0990/gotun/pkg/pool"
	"io"
	"time"
)

// buf cannot larger than 64k,because of the socket buffer size header is 16bit
const SocketBufSize = 20480
const MaxSegmentSize = 65535

func Pipe(left Stream, right Stream, timeout time.Duration) error {
	results := make(chan error, 2)
	defer close(results)

	go func() {
		_, err := unidirectionalStream(left, right, timeout)
		results <- err
		left.SetReadDeadline(time.Now())
	}()

	_, err := unidirectionalStream(right, left, timeout)
	results <- err
	right.SetReadDeadline(time.Now())

	first := <-results
	<-results
	return first
}

func unidirectionalStream(dst Stream, src Stream, timeout time.Duration) (written int64, err error) {
	buf := pool.GetBuf(SocketBufSize)
	defer pool.PutBuf(buf)

	for {
		if timeout != 0 {
			src.SetReadDeadline(time.Now().Add(timeout))
		}
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = errors.New("short write")
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
