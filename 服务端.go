package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

//ChatRoomManager 客户端管理
type ChatRoomManager struct {

	//储存并管理所有的连接client，在线的为true
	clients map[*ChatRoom]bool
	//接收web端发送来的的message，最后发给所有的client
	broadcast chan []byte
	//成功连接的client
	register chan *ChatRoom
	//注销的client
	unregister chan *ChatRoom
}

//ChatRoom 连接池
type ChatRoom struct {

	//成功连接的用户id
	clientID string
	//存放连接指针
	sock *websocket.Conn
	//发送的消息
	msg chan []byte
}

//Message 格式化成json，消息struct
type Message struct {

	//  发送者
	Sender string `json:"sender,omitempty"`
	//  内容
	Content string `json:"content,omitempty"`
}

//
func (manager *ChatRoomManager) start() {

	for {

		fmt.Println("开启start携程")
		select {
		//有新链接加入
		case conn := <-manager.register:
			fmt.Println("新连接")
			//有新链接加入，值为true
			manager.clients[conn] = true
			//把返回连接成功的消息json格式化
			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
			//调用客户端的send方法，发送消息
			manager.send(jsonMessage, conn)
		//如果连接断开
		case conn := <-manager.unregister:
			//判断连接的状态，如果是true，就关闭，删除连接client的值
			if _, ok := manager.clients[conn]; ok {
				close(conn.msg)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "/一位用户退出聊天室-_-!."})
				manager.send(jsonMessage, conn)
			}
		//发送消息给每个加入的客户端
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.msg <- message:
				default:
					close(conn.msg)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

// 遍历每个client发送信息
func (manager *ChatRoomManager) send(message []byte, ignore *ChatRoom) {

	for conn := range manager.clients {
		//不给屏蔽的连接发送消息
		if conn != ignore {

			//给conn.msg发消息，writeMSG函数收到调用conn.writemessage
			conn.msg <- message
		}
	}
}

//接收数据
func (cr *ChatRoom) writeMSG() {

	// 函数结束前关闭conn
	defer func() {
		cr.sock.Close()
	}()
	for {
		fmt.Println("启动writeMSG携程")
		select {
		case message, ok := <-cr.msg:
			fmt.Println("收到消息")
			if !ok {
				fmt.Println("没有消息")
				cr.sock.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//有消息就写入，发送给web端
			fmt.Println("发消息给web")
			cr.sock.WriteMessage(websocket.TextMessage, message)
		}
	}
}
func (cr *ChatRoom) readMSG() {

	defer func() {
		manager.unregister <- cr
		cr.sock.Close()
	}()

	for {
		fmt.Println("启动readMSG携程")
		//读取消息
		_, message, err := cr.sock.ReadMessage()
		//如果有错误信息，就注销这个连接然后关闭
		if err != nil {
			manager.unregister <- cr
			cr.sock.Close()
			break
		}
		//如果没有错误信息就把信息放入broadcast
		fmt.Println("服务端收到消息")
		jsonMessage, _ := json.Marshal(&Message{Sender: cr.clientID, Content: string(message)})
		manager.broadcast <- jsonMessage
	}
}

// 将http协议升级成websocket协议
// Upgrader指定 将HTTP连接升级到WebSocket连接 的参数
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 接收请求
func echo(w http.ResponseWriter, r *http.Request) {

	fmt.Println("我是echo处理器")
	// 通过Upgrader函数来建立一个websocket连接,返回一个Conn和一个error
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("upgrade错误:\n", err)
	}
	//每一次连接都会新开一个client，client.id通过uuid生成保证每次都是不同的
	client := &ChatRoom{
		// 用uuid生成唯一用户id
		clientID: uuid.Must(uuid.NewV4(), nil).String(),
		sock:     ws,
		msg:      make(chan []byte),
	}
	//注册一个新的链接
	manager.register <- client

	//开启携程接受web传来的消息
	go client.readMSG()
	//开启携程把消息发给web端
	go client.writeMSG()
}

func login(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method) //获取请求的方法
	if r.Method == "GET" {
		t, _ := template.ParseFiles("C:\\Users\\LEMon-X\\go\\public\\前端源码\\Login.html")
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		t, _ := template.ParseFiles("C:\\Users\\LEMon-X\\go\\public\\前端源码\\SimpleChat.html")
		t.Execute(w, nil)
	}
}

//创建连接入聊天室的管理者
var manager = ChatRoomManager{
	broadcast:  make(chan []byte),
	register:   make(chan *ChatRoom),
	unregister: make(chan *ChatRoom),
	clients:    make(map[*ChatRoom]bool),
}

func main() {

	log.SetFlags(0)
	go manager.start()
	fmt.Println("go start")
	// 处理器函数
	http.HandleFunc("/login", login)
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
