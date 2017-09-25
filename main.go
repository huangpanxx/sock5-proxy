package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
)

var (
	// AuthenticationResponse 初始响应的认证数据
	//AuthenticationResponse = []byte{0x05, 0x00}
	AuthenticationResponse = []byte{5, 0}
)

func handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	version, _ := reader.ReadByte()
	methodLen, _ := reader.ReadByte()
	method := make([]byte, methodLen)
	n, _ := reader.Read(method)

	fmt.Println(version, methodLen, method, n)

	if n, err := conn.Write(AuthenticationResponse); err != nil {
		fmt.Println(err, n)
		return
	}

	requestMata := make([]byte, 1024)
	n, err := reader.Read(requestMata)
	if err != nil {
		fmt.Println(err)
		return
	}
	remoteTye := requestMata[3]
	remoteAddr := requestMata[4:]
	remotePort := requestMata[n-2 : n]
	var host, port string
	switch remoteTye {
	case 0x01: // ipv4
		ipv4 := remoteAddr[:4]
		host = net.IPv4(ipv4[0], ipv4[1], ipv4[2], ipv4[3]).String()
	case 0x03: // domain
		domainLength := int(remoteAddr[0])
		host = string(remoteAddr[1 : domainLength+1])
		fmt.Println(domainLength, remoteAddr[1:domainLength+1])
	case 0x04: // ipv6
		ipv6 := remoteAddr[:16]
		host = net.IP{ipv6[0], ipv6[1], ipv6[2], ipv6[3], ipv6[4], ipv6[5], ipv6[6], ipv6[7], ipv6[8], ipv6[9], ipv6[10], ipv6[11], ipv6[12], ipv6[13], ipv6[14], ipv6[15]}.String()
	}

	port = strconv.Itoa(int(remotePort[0])<<8 | int(remotePort[1]))

	fmt.Println(host, port)

	server, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer server.Close()
	//conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
	conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})

	go io.Copy(server, conn)
	io.Copy(conn, server)
}

func main() {
	listen, err := net.Listen("tcp", ":1025")
	if err != nil {
		fmt.Println(err)
		panic("listen error")
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handle(conn)
	}

}