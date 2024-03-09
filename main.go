package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type FileServer struct {
}

func start(fs *FileServer) {
	event, err := net.Listen("tcp", ":3000")

	if err != nil {
		log.Fatal("Error -> ", err)
	}

	for {
		conn, err := event.Accept()

		if err != nil {
			log.Fatal("Error -> ", err)
		}

		go fs.readLoop(conn)
	}
}

func (fs *FileServer) readLoop(conn net.Conn) {
	buffer := make([]byte, 2048)
	for {
		n, err := conn.Read(buffer)

		if err != nil {
			log.Fatal("Error -> ", err)
		}

		file := buffer[:n]
		fmt.Println(file)
		fmt.Printf("%d bytes of data received over the network\n", n)
	}
}

func sendFile(size int) error {
	file := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, file)

	if err != nil {
		log.Fatal("Error -> ", err)
	}

	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal("Error -> ", err)
	}

	n, err := conn.Write(file)
	if err != nil {
		log.Fatal("Error -> ", err)
	}
	fmt.Printf("%d bytes of data written over the network\n", n)
	return nil
}

func main() {
	go func() {
		time.Sleep(4 * time.Second)
		sendFile(4000)
	}()

	server := &FileServer{}
	start(server)
}
