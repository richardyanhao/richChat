package chatroom

import (
	"log"
	"errors"
	"encoding/binary"
)

func (connectConn *WebSocketConn) ReadData() ([]byte, error){
	var frameContent [8]byte
	// 读取报文的第一、二个字节，第一个字节内容包括FIN\REV\OPCODE
	// 第二个字节内容有 mask bit\payload len
	if _, err := connectConn.conn.Read(frameContent[:2]); err != nil {
		return nil, err
	}
	// 根据第一个 Bit 判断是否是结束报文
	isFinal := frameContent[0] & finalFlag << 7 != 0
	if  !isFinal {
		log.Println("Recived fragmented frame, not support")
		return nil, errors.New("not support fragmented message")
	}

	/**
	*  获取 opcode，第一个字节第 5-8 位表达 opcode
	*  %x0 表示连续消息片断
	*  %x1 表示文本消息片断
	*  %x2 表未二进制消息片断
	*  %x3-7 为将来的非控制消息片断保留的操作码
	*  %x8 表示连接关闭
	*  %x9 表示心跳检查的ping
	*  %xA 表示心跳检查的pong
	*  %xB-F 为将来的控制消息片断的保留操作码
	*/
	opcodeValue := frameContent[0] & opcodeFlag
	if opcodeValue == CloseMessage {
		log.Println("Recived closed message, connection will be closed")
		return nil, errors.New("recived clode message")
	}
	if opcodeValue != TextMessage {
		return nil, errors.New("only support text message")
	}
	// 操作第二个字节,mask 如果设置为 1,掩码键必须放在 masking-key
	// 客户端发送给服务端的所有消息，此位的值都是 1
	isMasked := frameContent[1] & maskFlag != 0

	// 按照数据帧内容，处理 payload length
	payloadLength := int64(frameContent[1] & 0x7F)
	var dataLength int64
	// 如果 payload length <= 125  该值为数据长度
	// 如果 payload length = 126   后续 2 个字节表示数据长度
	// 如果 payload length = 127   后续 8 个字节表示数据长度
	switch payloadLength {
	case 126:
		if _, err := connectConn.conn.Read(frameContent[:2]); err != nil {
			return nil, err
		}
		dataLength = int64(binary.BigEndian.Uint16(frameContent[:2]))
	case 127:
		if _, err := connectConn.conn.Read(frameContent[:8]); err != nil {
			return nil, err
		}
		dataLength = int64(binary.BigEndian.Uint16(frameContent[:2]))
	default:
		dataLength = int64(payloadLength)
	}
	log.Printf("Read data length: %d, payload length %d", payloadLength, dataLength)

	// 读取 mask key
	if isMasked {
		if _, err := connectConn.conn.Read(connectConn.maskKey[:]); err != nil {
			return nil, err
		}
	}
	// 读取数据内容
	frameBody := make([]byte, dataLength)
	if _, err := connectConn.conn.Read(frameBody); err != nil {
		return nil, err
	}
	if isMasked {
		maskBytes(connectConn.maskKey, frameBody)
	}
	return frameBody, nil
}