// wsprobe 验证 gorilla/websocket 实时私信：A 监听，B 发送，A 收到推送。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type msg struct {
	Event string `json:"event"`
	Data  struct {
		ID      string `json:"id"`
		FromID  string `json:"from_id"`
		ToID    string `json:"to_id"`
		Content string `json:"content"`
	} `json:"data"`
}

func main() {
	wsURL := flag.String("url", "ws://localhost:3103/api/v1/chat/ws", "websocket url")
	tokenA := flag.String("token-a", "", "listener token")
	tokenB := flag.String("token-b", "", "sender token")
	toID := flag.String("to", "", "recipient user id for B's message")
	flag.Parse()
	if *tokenA == "" || *tokenB == "" {
		fmt.Println("usage: wsprobe -token-a <listener> -token-b <sender> -to <recipient>")
		os.Exit(2)
	}

	connA, _, err := websocket.DefaultDialer.Dial(*wsURL+"?token="+*tokenA, nil)
	if err != nil {
		fmt.Println("FAIL dial A:", err)
		os.Exit(1)
	}
	defer connA.Close()
	connB, _, err := websocket.DefaultDialer.Dial(*wsURL+"?token="+*tokenB, nil)
	if err != nil {
		fmt.Println("FAIL dial B:", err)
		os.Exit(1)
	}
	defer connB.Close()

	payload, _ := json.Marshal(map[string]string{"type": "message", "to_id": *toID, "msg_type": "text", "content": "ws-probe-hello"})
	if err := connB.WriteMessage(websocket.TextMessage, payload); err != nil {
		fmt.Println("FAIL write B:", err)
		os.Exit(1)
	}

	_ = connA.SetReadDeadline(time.Now().Add(8 * time.Second))
	_, raw, err := connA.ReadMessage()
	if err != nil {
		fmt.Println("FAIL read A:", err)
		os.Exit(1)
	}
	var m msg
	if err := json.Unmarshal(raw, &m); err != nil {
		fmt.Println("FAIL parse:", err, string(raw))
		os.Exit(1)
	}
	if m.Event != "message" || m.Data.Content != "ws-probe-hello" {
		fmt.Printf("FAIL unexpected payload: %s\n", string(raw))
		os.Exit(1)
	}
	fmt.Printf("PASS ws realtime message received from=%s to=%s content=%s\n", m.Data.FromID, m.Data.ToID, m.Data.Content)
}
