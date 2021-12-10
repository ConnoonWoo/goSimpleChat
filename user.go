/**
 *  @Description:
 *  @Author: Connoon.Wu
 *  @File: user
 *  @Version: 1.0.0
 *  @Date: 2021/12/8 15:45
 */

package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	Ch chan string
	conn net.Conn
	// 属于哪个server
	Server *Server
}

func NewUser(conn net.Conn,server *Server) *User {
	name := conn.RemoteAddr().String()
	user := &User{
		Name: name,
		Addr: name,
		Ch: make(chan string),
		conn: conn,
		Server: server,
	}

	 go user.ListenMessage()

	return user
}

// ListenMessage 监听消息，有消息就发送给客户端
func (u *User) ListenMessage() {
	for  {
		message := <-u.Ch
		u.conn.Write([]byte(message + "\n"))
	}
}

// Online 上线
func (u *User) Online() {
	u.Server.mapLock.Lock()
	u.Server.OnlineUser[u.Name] = u
	u.Server.mapLock.Unlock()

	// 广播当前用户上线了
	u.Server.BroadCast(u, "已上线")
}

// OffLine 下线
func (u *User) OffLine() {
	u.Server.mapLock.Lock()
	delete(u.Server.OnlineUser,u.Name)
	u.Server.mapLock.Unlock()

	// 广播当前用户下线
	u.Server.BroadCast(u, "已下线")
}

// DoMessage 正常发送消息
func (u *User) DoMessage(message string) {
	if message == "who" {
		// 查询在线用户
		u.Server.mapLock.Lock()
		for _, user := range u.Server.OnlineUser {
			message := "[" + user.Addr + "]" + user.Name + ": 在线....\n"
			u.sendMessageToSelf(message)
		}
		u.Server.mapLock.Unlock()
	} else if strings.HasPrefix(message,"rename|") {
		// 修改用户名称
		newName := strings.Split(message,"|")[1]
		if _,ok := u.Server.OnlineUser[newName]; ok {
			u.sendMessageToSelf("当前名称已被使用")
		} else {
			u.Server.mapLock.Lock()
			delete(u.Server.OnlineUser,u.Name)
			u.Server.OnlineUser[newName] = u
			u.Server.mapLock.Unlock()
			u.Name = newName
			u.sendMessageToSelf("修改用户名成功：" + newName + "\n")
		}
	} else if len(message) > 4 && strings.HasPrefix(message, "to|") {
		// 私聊  消息格式 to|谁|消息
		msgArr := strings.Split(message, "|")
		if msgArr[1] == "" || msgArr[2] == "" {
			u.sendMessageToSelf("消息格式有误，正确格式：to|里斯|你好")
			return
		}
		// 获取谁
		user,ok := u.Server.OnlineUser[msgArr[1]]
		if !ok {
			u.sendMessageToSelf("查无此人")
			return
		}
		// 写到对应channel
		//user.Ch <- msgArr[2]
		user.sendMessageToSelf(u.Name + "对你说：" + msgArr[2])
	} else {
		u.Server.BroadCast(u, message)
	}
}

// 给自己发消息
func (u *User) sendMessageToSelf(message string) {
	u.conn.Write([]byte(message))
}
