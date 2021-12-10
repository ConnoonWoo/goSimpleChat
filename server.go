/**
 *  @Description:
 *  @Author: Connoon.Wu
 *  @File: server
 *  @Version: 1.0.0
 *  @Date: 2021/12/8 15:28
 */

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip string
	Port int

	// 在线用户map
	OnlineUser map[string]*User
	mapLock sync.RWMutex
	// 广播消息的channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:   ip,
		Port: port,
		OnlineUser: make(map[string]*User),
		Message: make(chan string),
	}
}

func (s *Server) Start() {
	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		log.Fatal("socket listen error: ", err)
	}
	// socket close
	defer listen.Close()

	// 监听message
	go s.ListenMessager()

	for {
		// accept
		accept, err := listen.Accept()
		if err != nil {
			log.Fatal("accept error: ", err)
		}

		// do handler
		go s.Handler(accept)
	}
}

func (s *Server) Handler(accept net.Conn) {
	// 写入onlineMap
	user := NewUser(accept, s)
	user.Online()

	// 用户活跃状态
	isLive := make(chan bool)
	// 用户发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			read, err := accept.Read(buf)
			if read == 0 {
				user.OffLine()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err： ", err)
				return
			}

			user.DoMessage(string(buf[:read-1]))
			isLive <- true
		}
	}()

	// 当前handler阻塞 超时踢出群聊
	for {
		select {
		case <-isLive:
			// 活跃状态,重置定时器
			//不做任何事，激活select，更新下面定时器
		case <-time.After(time.Second * 50):
			// 超过秒数其他的case没有执行,超时
			user.sendMessageToSelf("你被踢了")
			// 关闭资源
			close(user.Ch)
			// 关闭链接
			accept.Close()
			// 退出handler
			return
		}
	}
}

// BroadCast 广播消息
func (s *Server) BroadCast(user *User,message string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + message
	s.Message <- sendMsg
}

// ListenMessager ListenMessage 监听消息channer,发送到所有在线用户得channer中
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message
		s.mapLock.Lock()
		for _, user := range s.OnlineUser {
			user.Ch <- msg
		}
		s.mapLock.Unlock()
	}
}
