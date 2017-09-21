package chatroom

import (
	"net"
	"net/http"
	"strings"
	"crypto/sha1"
	"encoding/base64"
	"log"
)

const finalFlag = 1
const maskFlag = 1 << 7
const TextMessage = 1
const CloseMessage = 8
const opcodeFlag = 0xf

var ConnMap = make(map[int]*WebSocketConn, 10)
var tempId = 1
var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

type WebSocketConn struct {
	writeBuf []byte
	maskKey  [4]byte
	conn net.Conn
}

func maskBytes(maskKey [4]byte, content []byte)  {
	pos := 0
	for i := range content {
		content[i] ^= maskKey[pos%4]
		pos++
	}
}

func newWebConn(newConn net.Conn) *WebSocketConn {
	newWebConn := &WebSocketConn{conn: newConn}
	ConnMap[tempId] = newWebConn
	tempId ++
	log.Println(ConnMap)
	return newWebConn
}

func (connectConn *WebSocketConn) Close()  {
	connectConn.conn.Close()
}

func isValueMatchInHead(header http.Header, paramName string, paramValue string) bool {
	for _, value := range header[paramName] {
		for _, s := range strings.Split(value, ",") {
			if strings.EqualFold(paramValue, strings.TrimSpace(s)) {
				return true
			}
		}
	}
	return false
}

func computeAcceptKey(challengeKey string) string{
	//产生一个散列值得方式是 sha1.New()
	h := sha1.New()
	//写入要处理的字节。如果是一个字符串，需要使用[]byte(s) 来强制转换成字节数组
	h.Write([]byte(challengeKey))
	h.Write(keyGUID)
	//sum 用来得到最终的散列值的字符切片
	return  base64.StdEncoding.EncodeToString(h.Sum(nil))
}
