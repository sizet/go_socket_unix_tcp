// ©.
// https://github.com/sizet/go_socket_unix_tcp

package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "net"
    "time"
)

const (
    // 使用 UNIX domain socket.
    unixNet = string("unix")
    // 要連線的 socket 路徑.
    unixRemoteAddr = string("/tmp/go.unixsocket")
)

// 是否結束程式, false = 否, true = 是.
var exitProcess bool




// 處理信號.
// 參數 :
// sigQueue
//   接收信號的 channel.
func signalHandle(
    sigQueue chan os.Signal) {

    var sigNum os.Signal

    for ;; {
        sigNum = <- sigQueue
        fmt.Printf("signal %d\n", sigNum)

        if (sigNum == syscall.SIGINT) ||
           (sigNum == syscall.SIGQUIT) ||
           (sigNum == syscall.SIGTERM) {
            exitProcess = true
        }
    }
}

// 處理傳送和接收資料.
// 參數 :
// remoteConnRD
//   和遠端建立的連線資源.
// 回傳 :
// fErr
//   是否發生錯誤, nil = 否, not nil = 是.
func remoteHandle(
    remoteConnRD *net.UnixConn)(
    fErr error) {

    var dataLen int
    var sendMsg string
    var recvBuf []byte
    var netDeadline time.Time

    // 設定傳送超時.
    netDeadline = time.Now().Add(3 * time.Second)
    fErr = remoteConnRD.SetDeadline(netDeadline)
    if fErr != nil {
        fmt.Printf("call net.SetDeadline(%v) fail [%s]\n", netDeadline, fErr.Error())
        return
    }
    // 填充要傳送的資料.
    sendMsg = fmt.Sprintf("aaa111223")
    fmt.Printf("send [%d][%s]\n", len(sendMsg), sendMsg)
    // 傳送資料.
    dataLen, fErr = remoteConnRD.Write([]byte(sendMsg))
    if fErr != nil {
        fmt.Printf("call net.Write(%s) fail [%s]\n", sendMsg, fErr.Error())
        return
    }
    if len(sendMsg) != dataLen {
        fmt.Printf("net.Write() len not match [%d][%d]\n", len(sendMsg), dataLen)
        return
    }

    // 設定接收超時.
    netDeadline = time.Now().Add(3 * time.Second)
    fErr = remoteConnRD.SetDeadline(netDeadline)
    if fErr != nil {
        fmt.Printf("call net.SetDeadline(%v) fail [%s]\n", netDeadline, fErr.Error())
        return
    }
    // 接收資料.
    recvBuf = make([]byte, 256)
    dataLen, fErr = remoteConnRD.Read(recvBuf)
    if fErr != nil {
        fmt.Printf("call net.Read() fail [%s]\n", fErr.Error())
        return
    }
    if dataLen == 0 {
        fmt.Printf("client connection close\n")
        return
    }
    // 顯示收到的資料.
    fmt.Printf("recv [%d][%s]\n", dataLen, string(recvBuf))

    return
}

func main() {

    var cErr error
    var exitCode int = -1
    var sigQueue chan os.Signal
    var remoteAddr net.Addr
    var remoteConnRD *net.UnixConn
    var tmpConnRD net.Conn
    var netTimeout time.Duration

    fmt.Printf("AF_UNIX TCP client, pid %d\n", os.Getpid())

    defer os.Exit(exitCode)

    // 設定信號處理方式.
    sigQueue = make(chan os.Signal, 4)
    go signalHandle(sigQueue)
    signal.Notify(sigQueue, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
    signal.Ignore(syscall.SIGPIPE)

    // 設定連線超時的時間 (3 秒).
    netTimeout = 3 * time.Second
    // 開始連線.
    tmpConnRD, cErr = net.DialTimeout(unixNet, unixRemoteAddr, netTimeout)
    if cErr != nil {
        fmt.Printf("call net.DialTimeout(%s, %s, %v) fail [%s]\n",
                   unixNet, unixRemoteAddr, netTimeout, cErr.Error())
        return
    }
    remoteConnRD = tmpConnRD.(*net.UnixConn)
    defer remoteConnRD.Close()

    // 顯示連接的 server 的位址.
    remoteAddr = remoteConnRD.RemoteAddr()
    fmt.Printf("connect [%s][%s]\n", remoteAddr.Network(), remoteAddr.String())

    // 處理傳送和接收資料.
    cErr = remoteHandle(remoteConnRD)
    if cErr != nil {
        fmt.Printf("call remoteHandle() fail\n")
        return
    }

    exitCode = 0

    return
}
