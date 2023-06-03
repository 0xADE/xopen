package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	waitOnConnect  = 1 * time.Second
	waitOnTransfer = 500 * time.Millisecond
)

func main() {
	//	ctx := context.Background()
	w := ui()

	log.Println("connect")
	conn, _ := net.DialTimeout("udp", "localhost:2782", waitOnConnect)
	go interact(conn)
	log.Println("wait signal")
	w.ShowAndRun()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc
	log.Println("exit")
}

func interact(conn net.Conn) {
	for {
		conn.SetDeadline(time.Now().Add(waitOnTransfer))
		log.Println("send command")
		conn.Write([]byte("list-exe"))
		buf := bufio.NewReader(conn)
		inp, err := buf.ReadString('\n')
		if err != nil {
			log.Printf("%s\n", err)
		}
		log.Println("read reply")
		data <- inp
		log.Println(inp)
	}
}

func copyTo(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}
