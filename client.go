/**
 *  @Description:
 *  @Author: Connoon.Wu
 *  @File: client
 *  @Version: 1.0.0
 *  @Date: 2021/12/9 16:48
 */
// todo 两点优化： 1.聊天空格会被拆分成两条传输 2.私聊怎么校验输入的是否是上面的用户

package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp string
	ServerPort int
	Conn net.Conn
	Name string
	flag int // 当前用户模式
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.chooseMenu() != true {
		}

		switch c.flag {
		case 1:
			c.publicChat()
			break
		case 2:
			c.privateChat()
			break
		case 3:
			c.updateName()
			break
		}

	}
}

func (c *Client) chooseMenu() bool {
	var flagNumber int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.修改用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flagNumber)
	if flagNumber >= 0 && flagNumber <= 3 {
		c.flag = flagNumber
		return true
	} else {
		fmt.Println("------请选择正确的数字")
		return false
	}
}

func (c *Client) publicChat() {
	var chat string
	fmt.Println(">>>>>请输入聊天内容，exit退出")
	fmt.Scanln(&chat)

	for chat != "exit" {
		if len(chat) != 0 {
			sendMsg := chat + "\n"
			_, err := c.Conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("Conn write error:" , err)
				break
			}
		}

		chat = ""
		fmt.Println(">>>>>请输入聊天内容，exit退出")
		fmt.Scanln(&chat)
	}
}

func (c *Client) privateChat() {
	var remoteName string
	var msg string
	c.selectUsers()

	fmt.Println("<<<<<请选择私聊用户，exit退出")
	fmt.Scanln(&remoteName)
	if remoteName == c.Name {
		c.selectUsers()
		fmt.Println("<<<<<请选择非自己的私聊用户，exit退出")
		fmt.Scanln(&remoteName)
	}
	// todo 还可以校验输入的用户名存不存在
	for remoteName != "exit" {
		fmt.Println("<<<<<请输入聊天内容，exit退出")
		fmt.Scanln(&msg)

		for msg != "exit" {
			if len(msg) != 0 {
				sendMsg := "to|" + remoteName + "|" + msg + "\n\n"
				_, err := c.Conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("Conn write error:" , err)
					break
				}
			}

			msg = ""
			fmt.Println(">>>>>请输入聊天内容，exit退出")
			fmt.Scanln(&msg)
		}

		c.selectUsers()
		fmt.Println("<<<<<请选择私聊用户，exit退出")
		fmt.Scanln(&remoteName)
		if remoteName == c.Name {
			c.selectUsers()
			fmt.Println("<<<<<请选择非自己的私聊用户，exit退出")
			fmt.Scanln(&remoteName)
		}
	}

}

func (c *Client) selectUsers() {
	sendMsg := "who" + "\n"
	_, err := c.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("获取在线用户错误")
		return
	}
}

func (c *Client) updateName() bool {
	fmt.Println(">>>>>>输入用户名：")
	fmt.Scanln(&c.Name)
	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error : ", err)
		return false
	}

	return true
}

// 处理服务端返回的消息
func (c *Client) dealResponse() {
	// 永久堵塞监听
	io.Copy(os.Stdout,c.Conn)
}

func NewClient(ip string, port int) *Client {
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		flag: 999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {
		fmt.Println("dial error:", err)
		return nil
	}

	client.Conn = conn
	return client
}

var (
	serverIp string
	serverPort int
)

func init() {
	flag.StringVar(&serverIp,"ip","127.0.0.1","设置服务器地址（默认：127.0.0.1）")
	flag.IntVar(&serverPort,"port",9999,"设置服务器端口（默认：9999）")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("------连接服务器失败------")
		return
	}

	// 单独开启goroutine 处理server的回执消息
	go client.dealResponse()

	fmt.Println("------连接服务器成功------")

	client.Run()
}
