package conn_wrapper

import (
	"encoding/binary"
	"net"
)

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

type ConnWrapper struct {
	Conn net.Conn
}

func NewConn(c net.Conn) *ConnWrapper {
	return &ConnWrapper{
		Conn: c,
	}
}

func (c *ConnWrapper) Read() (msgType byte, msg []byte, err error) {
	buf := make([]byte, 1024)
	readLen, err := c.Conn.Read(buf)
	if err != nil {
		return MsgTypeUndefined, nil, err
	}
	if readLen < 5 {
		return MsgTypeUndefined, nil, cwError{"Too short msg"}
	}
	msgLen := int(binary.BigEndian.Uint32(buf[0:4]))
	msgType = buf[4]
	msg = append(msg, buf[5:readLen]...)
	for readLen < msgLen {
		len, err := c.Conn.Read(buf)
		if err != nil {
			return MsgTypeUndefined, nil, err
		}
		msg = append(msg, buf[0:readLen]...)
		readLen += len
	}
	if readLen > msgLen {
		return MsgTypeUndefined, nil, cwError{"Too long msg"}
	}
	return // All return values are already set
}

func (c *ConnWrapper) Write(msgType byte, msg []byte) error {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, uint32(len(msg)+5))
	buf[4] = msgType
	buf = append(buf, msg...) // May be it's too expensive?
	l, err := c.Conn.Write(buf)
	if err != nil {
		return err
	}
	if l != len(buf) { // When can it happen
		return cwError{"Written len is not equal msg len"}
	}
	return nil
}
