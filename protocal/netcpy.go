package protocal

import (
	"io"
	"sync"
)

func netReadFromO(src io.Reader, dst io.Writer) (written int64, err error) {

	// 优化缓冲区大小
	size := 32 * 1024
	if lr, ok := src.(*io.LimitedReader); ok && lr.N < int64(size) {
		if lr.N > 0 {
			size = int(lr.N)
		} else {
			size = 1
		}
	}

	buf := make([]byte, size)
	readBuf := make([]byte, 0, size*2) // 双倍缓冲用于处理数据拼接
	var (
		contentLen   int
		headerParsed bool
	)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			readBuf = append(readBuf, buf[:nr]...)
		}

		// 处理缓冲数据
		for len(readBuf) > 0 {
			if !headerParsed {
				if err = parseHeader(&readBuf, &contentLen); err != nil {
					return written, err
				}
				if contentLen == 0 {
					continue
				}
				headerParsed = true
			}

			// 处理有效载荷
			chunk := readBuf[:min(contentLen, len(readBuf))]
			nw, ew := dst.Write(chunk)
			if nw < 0 {
				return written, ErrInvalidWrite
			}
			written += int64(nw)
			if ew != nil {
				return written, ew
			}

			// 更新状态
			contentLen -= nw
			readBuf = readBuf[nw:]

			// 当前数据包处理完成
			if contentLen == 0 {
				headerParsed = false
				if len(readBuf) > 0 && readBuf[0] == byte(Header_EOF) {
					return written, nil
				}
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

func parseHeader(buf *[]byte, contentLen *int) error {
	if len(*buf) < 1 {
		return nil
	}

	switch (*buf)[0] {
	case byte(Header_PAYLOAD):
		if len(*buf) < 5 {
			return nil // 头部不完整
		}
		*contentLen = int((*buf)[1])<<24 | int((*buf)[2])<<16 |
			int((*buf)[3])<<8 | int((*buf)[4])
		*buf = (*buf)[5:] // 消耗头部字节
		return nil

	case byte(Header_EOF):
		*buf = (*buf)[1:] // 消耗EOF字节
		return io.EOF

	default:
		return ErrInvalidRead
	}
}

var bufPool = sync.Pool{
	New: func() any {
		return make([]byte, 32*1024+5)
	},
}

func netWriteToO(src io.Reader, dst io.Writer) (written int64, err error) {
	// 从池中获取缓冲
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	// 动态调整缓冲区大小
	maxPayload := len(buf) - 5
	if lr, ok := src.(*io.LimitedReader); ok {
		if remaining := int(lr.N); remaining < maxPayload {
			if remaining > 0 {
				maxPayload = remaining
			} else {
				maxPayload = 1
			}
		}
	}

	var (
		eofSent    bool
		headerBuf  = buf[:5]
		payloadBuf = buf[5 : 5+maxPayload]
	)

	for !eofSent {
		// 读取数据块
		nr, er := src.Read(payloadBuf)
		if nr > 0 {
			// 构造协议头
			writeHeader(headerBuf, Header_PAYLOAD, nr)

			// 写入完整数据包（头+负载）
			if err = writeFull(dst, headerBuf, payloadBuf[:nr]); err != nil {
				return written, err
			}
			written += int64(nr)
		}

		// 处理结束条件
		if er != nil {
			if er != io.EOF {
				return written, er
			}

			// 发送EOF标记
			writeHeader(headerBuf, Header_EOF, 0)
			if _, err = dst.Write(headerBuf[:1]); err != nil {
				return written, err
			}
			eofSent = true
		}
	}

	return written, nil
}

// 协议头构造工具函数
func writeHeader(buf []byte, header Header, length int) {
	buf[0] = byte(header)
	if header == Header_PAYLOAD {
		buf[1] = byte(length >> 24)
		buf[2] = byte(length >> 16)
		buf[3] = byte(length >> 8)
		buf[4] = byte(length)
	}
}

// 可靠写入函数
func writeFull(w io.Writer, header, payload []byte) error {
	// 合并写入减少系统调用
	full := append(header, payload...)
	for len(full) > 0 {
		n, err := w.Write(full)
		if n < 0 {
			return ErrInvalidWrite
		}
		if err != nil {
			return err
		}
		full = full[n:]
	}
	return nil
}
