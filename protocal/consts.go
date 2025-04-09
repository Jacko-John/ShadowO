package protocal

import "errors"

type Header byte

const (
	Header_EOF     Header = 0x01
	Header_PAYLOAD Header = 0x02
	Header_NEW     Header = 0x03
	Header_GOT     Header = 0x04
	Header_PING    Header = 0x05
	Header_PONG    Header = 0x06
	Header_AUTH    Header = 0x07
	Header_AUTH_R  Header = 0x08
)

type Identity byte

const (
	ID_MASTER Identity = 0x01
	ID_TUNNEL Identity = 0x02
)

var (
	ErrAuthFailed   = errors.New("auth failed")
	ErrAuthHeader   = errors.New("auth header error")
	ErrInvalidWrite = errors.New("invalid write result")
	ErrInvalidRead  = errors.New("invalid read result")
	ErrPong         = errors.New("pong error")
	ErrPing         = errors.New("ping error")
	ErrNew          = errors.New("new error")
	ErrGot          = errors.New("got error")
	ErrEOF          = errors.New("eof error")
)
