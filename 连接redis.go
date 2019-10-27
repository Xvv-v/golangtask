package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/garyburd/redigo/redis"
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
func (manage *ChatRoomManager) start() {

	for {

		select {
		//有新链接加入
		case conn := <-manage.register:
			//有新链接加入，值为true
			manage.clients[conn] = true
			//有新连接加入时，就显示历史记录
			conn.GetRedisMsg(conn.sock)
		//如果连接断开
		case conn := <-manage.unregister:
			//判断连接的状态，如果是true，就关闭，删除连接client的值
			if _, ok := manager.clients[conn]; ok {
				close(conn.msg)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "/一位用户退出聊天室-_-!."})
				manager.send(jsonMessage, conn)
			}
		//发送消息给每个加入的客户端
		case message := <-manager.broadcast:
			//遍历已经连接的客户端，把消息发送给他们
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
func (manage *ChatRoomManager) send(message []byte, ignore *ChatRoom) {

	for conn := range manager.clients {
		//不给屏蔽的连接发送消息
		if conn != ignore {
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
		select {
		case message, ok := <-cr.msg:
			if !ok {
				cr.sock.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//有消息就写入，发送给web端
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
		//读取消息
		_, message, err := cr.sock.ReadMessage()
		//如果有错误信息，就注销这个连接然后关闭
		if err != nil {
			manager.unregister <- cr
			cr.sock.Close()
			break
		}
		//如果没有错误信息就把信息放入broadcast
		jsonMessage, _ := json.Marshal(&Message{Sender: cr.clientID, Content: string(message)})
		manager.broadcast <- jsonMessage
		//把这条信息存入redis
		cr.SetRedisMsg(string(message))
	}
}

//SetRedisMsg 将消息写入redis
func (cr *ChatRoom) SetRedisMsg(msg string) {

	redisNumber++
	_, err := redisConn.Do("setex", strconv.Itoa(redisNumber), 300, msg) //将消息的条数作为key
	if err != nil {
		log.Println("存入信息出错：\n", err)
		return
	}
}

//GetRedisMsg 从redis中读出消息
func (cr *ChatRoom) GetRedisMsg(ws *websocket.Conn) {

	var msg string
	var err error
	for i := 0; i < redisNumber; i++ {

		msg, err = redis.String(redisConn.Do("get", strconv.Itoa(redisNumber)))
		if err != nil {

			log.Println("读取信息错误：\n", err)
		} else {
			var message Message
			//将json字符串转化为Message结构
			data := []byte(msg)
			json.Unmarshal(data, &message)
			cr.msg <- data
		}
	}
}

// 将http协议升级成websocket协议
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 接收请求
func echo(w http.ResponseWriter, r *http.Request) {

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
		t, _ := template.ParseFiles("C:\\Users\\LEMon-X\\go\\public\\前端源码\\RedisChat.html")
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
var redisConn redis.Conn

// redis储存的消息数
var redisNumber int

func main() {

	log.SetFlags(0)
	//连接redis
	var redisErr error
	redisConn, redisErr = redis.Dial("tcp", "127.0.0.1:6379")
	if redisErr != nil {

		log.Fatal("连接失败：", redisErr)
	} else {

		fmt.Println("连接成功")
	}
	// 关闭连接
	defer redisConn.Close()

	go manager.start()
	// 处理器函数
	http.HandleFunc("/login", login)
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
