package socks5

import (
	"errors"
	"github.com/0990/socks5/pkg/pool"
	"io"
	"sync/atomic"
	"time"
)

// buf cannot larger than 64k,because of the socket buffer size header is 16bit
const SocketBufSize = 20480
const MaxSegmentSize = 65535

func Pipe(left Stream, right Stream, timeout time.Duration) error {
	// 使用一个原子变量记录最近一次数据活动的时间（UnixNano格式）
	var lastActivity int64 = time.Now().UnixNano()

	// 辅助函数：更新活动时间
	updateActivity := func() {
		atomic.StoreInt64(&lastActivity, time.Now().UnixNano())
	}

	// 启动全局监控协程
	stopMonitor := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				last := atomic.LoadInt64(&lastActivity)
				if time.Since(time.Unix(0, last)) > timeout {
					// 当全局无活动超过timeout，关闭连接
					left.SetReadDeadline(time.Now())
					right.SetReadDeadline(time.Now())
					return
				}
			case <-stopMonitor:
				return
			}
		}
	}()

	// 启动双向转发
	results := make(chan error, 2)
	go func() {
		_, err := unidirectionalStream(left, right, updateActivity)
		results <- err
	}()
	_, err := unidirectionalStream(right, left, updateActivity)
	results <- err

	// 停止监控
	close(stopMonitor)

	// 只返回第一个出错的结果
	first := <-results
	<-results
	return first
}

// unidirectionalStream 将数据从 src 拷贝到 dst, 每次拷贝数据时调用 activityCallback 通知活动
func unidirectionalStream(dst Stream, src Stream, activityCallback func()) (written int64, err error) {
	buf := pool.GetBuf(SocketBufSize)
	defer pool.PutBuf(buf)

	for {
		// 这里不设置独立的超时，转而依靠全局监控
		nr, er := src.Read(buf)
		if nr > 0 {
			// 数据到达，更新活动时间
			activityCallback()
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
