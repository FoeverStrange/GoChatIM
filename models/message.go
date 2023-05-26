package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

// 消息
type Message struct {
	gorm.Model
	//发送者
	FromId int64
	//接收者
	TargetId int64
	//消息类型 1私聊、2群聊、3广播
	Type int
	//消息类型 1文字、2表情包3图片、4音频
	Media int
	//消息内容
	Context string
	Pic     string
	Url     string
	Desc    string
	//其他数字统计
	Amount int `json:"amount"`
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// TODO:链接的时候需要鉴权，即先登录再连接
// 需要：发送者ID、接收者ID、消息类型、发送的内容、发送类型
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.获取参数并校验合法性
	//检验token
	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	//token := query.Get("token")   //token校验
	//targetId := query.Get("targetId")
	//context := query.Get("context")
	//msgtype := query.Get("type")
	isvalid := true //TODO: checkToken()
	//2.获取conn
	conn, err := (&websocket.Upgrader{
		//用于token校验
		CheckOrigin: func(r *http.Request) bool {
			return isvalid
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}
	//3. 用户关系
	//4. userId根node绑定并且加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//5.完成发送的逻辑
	go sendProc(node)

	//6.完成接收的逻辑，从接口conn的管道中读取数据
	go recvProc(node)

	sendMsg(userId, []byte("欢迎进入聊天室"))
}

// 发送逻辑，把消息从DataQue推送到conn里
func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			//发给自己的管道口
			fmt.Printf("[ws] SendProc >>> %s\n", data)
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println("没写进去", err)
				return
			}
		}
	}
}

// 接收逻辑，从conn里接收信息
func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(data)
		broadMsg(data)
		// fmt.Printf("[ws] RecvProc <<<< %s\n", data)

	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(10, 104, 0, 255),
		Port: 3000,
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case data := <-udpsendChan:
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// udp数据接收协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer con.Close()
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑,对于一个消息，逻辑上是先在自己端回显，再转发给其他用户/群组去
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Printf("msg_Type: %d, TargetId: %d\n", msg.Type, msg.TargetId)
	switch msg.Type {
	case 1: //私信
		sendMsg(msg.TargetId, data)
		// case 2:  //群发
		// 	sendGroupMsg()
		// case 3:  //广播
		// 	sendAllMsg()
		// case 4:

	}
}

// 将msg推送到node的DataQue中
func sendMsg(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	if ok {
		node.DataQueue <- msg
	}
}
