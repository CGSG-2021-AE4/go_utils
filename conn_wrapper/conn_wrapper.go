package conn_wrapper

import (
	"encoding/binary"
	"fmt"
	"net"
)

// Errors

type cwError struct {
	err string
}

func (e cwError) Error() string {
	return e.err
}

const (
	MsgTypeUndefined    = iota // if error while reading
	MsgTypeError        = iota
	MsgTypeOk           = iota
	MsgTypeRegistration = iota
	MsgTypeRequest      = iota
	MsgTypeResponse     = iota
	MsgTypeClose        = iota
)

func FormatError(err byte) string {
	switch err {
	case MsgTypeUndefined:
		return "Undefined"
	case MsgTypeError:
		return "Error"
	case MsgTypeOk:
		return "Ok"
	case MsgTypeRegistration:
		return "Registration"
	case MsgTypeRequest:
		return "Request"
	case MsgTypeResponse:
		return "Response"
	case MsgTypeClose:
		return "Close"
	default:
		return "Invalid msg type"
	}
}

// Wrapper itself

type Conn struct {
	NetConn net.Conn
}

func New(c net.Conn) *Conn {
	return &Conn{
		NetConn: c,
	}
}

func (c *Conn) Read() (msgType byte, msg []byte, err error) {
	buf := make([]byte, 1024)
	readLen, err := c.NetConn.Read(buf)
	if err != nil {
		return MsgTypeUndefined, nil, err
	}
	msg = append(msg, buf[5:readLen]...)
	if readLen < 5 {
		return MsgTypeUndefined, msg, cwError{"Too short msg"}
	}
	msgLen := int(binary.BigEndian.Uint32(buf[0:4]))
	msgType = buf[4]
	for readLen < msgLen {
		len, err := c.NetConn.Read(buf)
		if err != nil {
			return MsgTypeUndefined, msg, err
		}
		msg = append(msg, buf[0:readLen]...)
		readLen += len
	}
	if readLen > msgLen {
		return MsgTypeUndefined, nil, cwError{fmt.Sprintf("Too long msg %d %d", readLen, msgLen)}
	}
	return // All return values are already set
}

func (c *Conn) Write(msgType byte, msg []byte) error {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, uint32(len(msg)+5))
	buf[4] = msgType
	buf = append(buf, msg...) // May be it's too expensive?
	l, err := c.NetConn.Write(buf)
	if err != nil {
		return err
	}
	if l != len(buf) { // When can it happen
		return cwError{"Written len is not equal msg len"}
	}
	return nil
}
