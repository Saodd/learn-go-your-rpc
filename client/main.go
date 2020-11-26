package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

type ConnWorker struct {
	conn net.Conn
	r    []byte
	w    []byte
}

var ConnPool = sync.Pool{New: func() interface{} {
	return &ConnWorker{r: make([]byte, 255), w: make([]byte, 0, 255)}
}}

func RemoteCall(address string, action byte, param interface{}) (resp []byte) {
	worker := ConnPool.Get().(*ConnWorker)
	defer ConnPool.Put(worker)

	for retry := 0; retry < 1; retry++ {
		if worker.conn == nil {
			conn, err := net.Dial("tcp", address)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			worker.conn = conn
		}

		var conn, r, w = worker.conn, worker.r, worker.w
		// 构造请求体
		// 第1位：消息总长度。第2位：动作。第3~位：参数json
		js, err := json.Marshal(param)
		if err != nil {
			log.Println(err)
		}
		w = append(w, byte(len(js)), action)
		w = append(w, js...)
		conn.Write(w)
		w = w[:0]

		// 读取响应体
		// 第1位：长度。第2~位：结果json
		_, err = conn.Read(r[:1])
		if err != nil {
			fmt.Println(err)
			conn.Close()
			conn = nil
			continue
		}
		n := int(r[0])
		_, err = conn.Read(r[:n])
		if err != nil {
			fmt.Println(err)
			conn.Close()
			conn = nil
			continue
		}
		return r[:n]
	}
	return nil
}

func main() {
	for i := 0; i < 50; i++ {
		resp := RemoteCall("localhost:8080", 1, ActionParam1{3, 6})
		fmt.Println("返回值：", string(resp))
	}
}

type ActionParam1 struct {
	A int `json:"a"`
	B int `json:"b"`
}
