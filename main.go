package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
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
	buffer := new(bytes.Buffer)
	for {
		var size int64
		binary.Read(conn, binary.LittleEndian, &size)
		n, err := io.CopyN(buffer, conn, size)

		if err != nil {
			log.Fatal("Error -> ", err)
		}

		fmt.Println(buffer.Bytes())
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

	binary.Write(conn, binary.LittleEndian, int64(size))

	n, err := io.CopyN(conn, bytes.NewReader(file), int64(size))
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
