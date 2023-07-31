package app

import (
	"bufio"
	"fmt"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestOS(t *testing.T) {
	fmt.Println(runtime.GOOS)
	fmt.Println(runtime.GOARCH)
	fmt.Println(runtime.Compiler)

	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		panic(err)
	}
	sysname := unix.ByteSliceToString(uts.Sysname[:])
	release := unix.ByteSliceToString(uts.Release[:])
	version := unix.ByteSliceToString(uts.Version[:])
	machine := unix.ByteSliceToString(uts.Machine[:])
	nodeName := unix.ByteSliceToString(uts.Nodename[:])
	fmt.Printf("sysname: %s\nrelease: %s\n", sysname, release)
	fmt.Printf("version: %s\nmachine: %s\n", version, machine)
	fmt.Printf("nodeName: %s\n", nodeName)
	if sysname == "Darwin" {
		dotPos := strings.Index(release, ".")
		if dotPos == -1 {
			fmt.Printf("invalid release value: %s\n", release)
			return
		}
		major := release[:dotPos]
		majorVersion, err := strconv.Atoi(major)
		if err != nil {
			fmt.Printf("invalid release value: %s, %v\n", release, err)
			return
		}
		fmt.Println("macOS >= Big Sur:", majorVersion >= 20)
	}
}

func TestSring(t *testing.T) {
	fmt.Println(generator.CamelCase("dfasSDFxadfx"))
}

func TestSocket(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		fmt.Println("client dial err=", err)
		return
	}
	defer conn.Close()
	for {
		fmt.Println("请输入信息，回车结束输入")
		reader := bufio.NewReader(os.Stdin)
		//终端读取用户回车，并准备发送给服务器
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("readString err=", err)
		}
		line = strings.Trim(line, "\r\n")
		if line == "exit" {
			fmt.Println("客户端退出...")
			break
		}
		line = strings.TrimSpace(line)
		//将line发送给服务器
		n, err := conn.Write([]byte(line))
		if err != nil {
			fmt.Println("conn.Write err=", err)
		}
		fmt.Printf("发送的内容了%d文字\n", n)
	}
}
