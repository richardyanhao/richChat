package chatroom

import (
	"encoding/binary"
)

func BroadcastMessage(data []byte) {
	for _, conn := range ConnMap {
		conn.WriteData(data)
	}
}

func (connectConn *WebSocketConn) WriteData(data []byte) {
	length := len(data)
	connectConn.writeBuf = make([]byte, length + 10)
	// 数据帧的第一个字节, 不支持分片，且值能发送文本类型数据
	// 所以二进制位为 %b1000 0001
	// b0 := []byte{0x81}
	connectConn.writeBuf[0] = finalFlag << 7 | byte(TextMessage)
	// 数据开始位置
	payloadStart := 2
	// 数据帧第二个字节 计算长度
	switch {
	case length >= 65536:
		connectConn.writeBuf[1] = byte(127)
		binary.BigEndian.PutUint64(connectConn.writeBuf[payloadStart:], uint64(length))
		// 需要 8 byte 存储数据长度
		payloadStart += 8
	case length > 125:
		connectConn.writeBuf[1] = byte(126)
		binary.BigEndian.PutUint64(connectConn.writeBuf[payloadStart:], uint64(length))
		// 需要两个字节存储数据长度
		payloadStart += 2
	default:
		connectConn.writeBuf[1] = byte(length)
	}
	copy(connectConn.writeBuf[payloadStart:], data[:])
	connectConn.conn.Write(connectConn.writeBuf[:payloadStart+length])
}