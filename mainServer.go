/**
 *  @Description:
 *  @Author: Connoon.Wu
 *  @File: mainServer
 *  @Version: 1.0.0
 *  @Date: 2021/12/8 15:38
 */

package main

func main() {
	server := NewServer("127.0.0.1", 9999)
	server.Start()
}
