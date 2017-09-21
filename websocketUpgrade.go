package chatroom

import (
	"net/http"
	"errors"
	"bufio"
	"net"
	"log"
)

func Upgrade(w http.ResponseWriter, r *http.Request) (*WebSocketConn, error) {
	// 校验 method
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return nil, errors.New("method not GET")
	}
	// 校验 Sec-WebSocket-Version 版本
	if value := r.Header.Get("Sec-WebSocket-Version"); value != "13" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return nil, errors.New("version != 13")
	}
	// 检查 Connection 是否是 Upgrade
	if !isValueMatchInHead(r.Header, "Connection", "Upgrade") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return nil, errors.New("connection is not upgrade")
	}

	challengeKey := r.Header.Get("Sec-Websocket-Key")
	if challengeKey == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return nil, errors.New("websocket: key missing or blank")
	}
	log.Println("If conn exist return conn")

	// 劫持
	// After a call to Hijack(), the HTTP server library
	// will not do anything else with the connection.
	// It becomes the caller's responsibility to manage
	// and close the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return nil, errors.New("websocket: response dose not implement http.Hijacker")
	}
	var (
		rw 		*bufio.ReadWriter
		netConn net.Conn
		br      *bufio.Reader
	)

	netConn, rw, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return nil, err
	}
	br = rw.Reader
	if br.Buffered() > 0 {
		netConn.Close()
		return nil, errors.New("websocket: client sent data before handshake is complete")
	}
	// 构造握手成功后返回的 response
	p := []byte{}
	p = append(p, "HTTP/1.1 101 Switching Protocals\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept:"...)
	p = append(p, computeAcceptKey(challengeKey)...)
	p = append(p, "\r\n\r\n"...)
	if _, err = netConn.Write(p); err != nil {
		netConn.Close()
		return nil, err
	}
	log.Println("Upgrade http to websocket successfully")
	conn := newWebConn(netConn)
	return conn, nil
}