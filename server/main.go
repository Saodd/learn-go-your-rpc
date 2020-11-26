package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
)

var logger = log.New(os.Stdout, "", log.Llongfile)

func main() {
	server, _ := net.Listen("tcp", ":8080")
	for {
		conn, err := server.Accept()
		if err != nil {
			logger.Fatalln(err)
		}
		go serve(conn)
	}
}

func serve(conn net.Conn) {
	logger.Println("连接：", conn.RemoteAddr())
	defer conn.Close()
	r := make([]byte, 255)
	w := make([]byte, 0, 255)
	for {
		// 解析请求体
		// 第1位：消息总长度。第2位：动作。第3~位：参数json
		if _, err := conn.Read(r[:2]); err != nil {
			logger.Println(err)
			return
		}
		var n, action = int(r[0]), r[1]
		if _, err := conn.Read(r[:n]); err != nil {
			logger.Println(err)
			return
		}
		// 执行服务端任务函数
		var js []byte
		js, err := route(action, r[:n])
		if err != nil {
			logger.Println(err)
			return
		}
		// 写回响应体
		w = w[:1]
		w[0] = byte(len(js))
		w = append(w, js...)
		if _, err := conn.Write(w); err != nil {
			logger.Println(err)
			return
		}
	}
}

func route(action byte, req []byte) ([]byte, error) {
	switch action {
	case 1:
		return processAction1(req)
	default:
		return nil, errors.New("不支持的action")
	}
}

type ActionParam1 struct {
	A int `json:"a"`
	B int `json:"b"`
}

type ActionResult1 struct {
	Sum int `json:"sum"`
}

func processAction1(req []byte) ([]byte, error) {
	var param ActionParam1
	err := json.Unmarshal(req, &param)
	if err != nil {
		return nil, err
	}

	// 构造响应体
	// 第1位：长度。第2~位：结果json
	var res ActionResult1
	res.Sum = param.A + param.B
	js, _ := json.Marshal(&res)
	return js, nil
}
