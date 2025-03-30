package protocal

import (
	"io"
)

// before deepseek optimization
func __netReadFromO(src io.Reader, dst io.Writer) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rf, ok := dst.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	buf := make([]byte, size)
	var l int
	var contentLen int = 0
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			if contentLen == 0 {
				if buf[0] == byte(Header_PAYLOAD) {
					if nr >= 5 {
						contentLen = int(buf[1])<<24 | int(buf[2])<<16 | int(buf[3])<<8 | int(buf[4])
						contentLen -= nr - 5
					} else {
						tmp := make([]byte, 5-nr)
						r, e := src.Read(tmp)
						if r > 0 {
							copy(buf[nr:nr+r], tmp[0:r])
							nr += r
							contentLen = int(buf[1])<<24 | int(buf[2])<<16 | int(buf[3])<<8 | int(buf[4])
						}
						if e != nil {
							err = e
							break
						}
					}
					l = 5
				} else if buf[0] == byte(Header_EOF) {
					break
				} else {
					err = ErrInvalidRead
					break
				}
			} else {
				l = 0
				contentLen -= nr
				if contentLen < 0 {
					prelen := contentLen + nr
					if buf[prelen] == byte(Header_PAYLOAD) {
						if contentLen <= -5 {
							contentLen = int(buf[prelen+1])<<24 | int(buf[prelen+2])<<16 | int(buf[prelen+3])<<8 | int(buf[prelen+4])
							contentLen -= (nr - prelen - 5)
							copy(buf[5:prelen+5], buf[0:prelen])
							l = 5
						} else {
							tmp := make([]byte, 5+contentLen)
							r, e := src.Read(tmp)
							if r > 0 {
								copy(buf[nr:nr+r], tmp[0:r])
								nr = prelen
								contentLen = int(buf[prelen+1])<<24 | int(buf[prelen+2])<<16 | int(buf[prelen+3])<<8 | int(buf[prelen+4])
							}
							if e != nil {
								err = e
								break
							}
						}
					} else if buf[0] == byte(Header_EOF) {
						w, e := dst.Write(buf[0:prelen])
						if w < 0 || prelen < w {
							w = 0
							if e == nil {
								e = ErrInvalidWrite
							}
						}
						written += int64(w)
						if e != nil {
							err = e
							break
						}
						break
					} else {
						err = ErrInvalidRead
						break
					}
				}
			}
			nw, ew := dst.Write(buf[l:nr])
			if nw < 0 || nr-l < nw {
				nw = 0
				if ew == nil {
					ew = ErrInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr-l != nw {
				err = io.ErrShortWrite
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

// after deepseek optimization
func __netWriteToO(src io.Reader, dst io.Writer) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rf, ok := dst.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	buf := make([]byte, size)
	var contentLen uint32 = 0
	for {
		nr, er := src.Read(buf[5:])
		if nr > 0 {
			contentLen = uint32(nr)
			buf[0] = byte(Header_PAYLOAD)
			buf[1] = byte(contentLen >> 24)
			buf[2] = byte(contentLen >> 16)
			buf[3] = byte(contentLen >> 8)
			buf[4] = byte(contentLen)
			nw, ew := dst.Write(buf[:5+nr])
			if nw < 0 || 5+nr < nw {
				nw = 0
				if ew == nil {
					ew = ErrInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if 5+nr != nw {
				err = io.ErrShortWrite
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
