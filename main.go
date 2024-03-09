package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type FileServer struct {
	shutdownSignal chan os.Signal
	waitGroup      sync.WaitGroup
}

func handleError(err error) {
	if err != nil {
		log.Println("Error ->", err)
	}
}

func (fs *FileServer) handleSignals() {
	signal.Notify(fs.shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a signal to gracefully shutdown
	<-fs.shutdownSignal

	// Notify goroutines to finish
	fs.waitGroup.Done()
}

func start(fs *FileServer) {
	event, err := net.Listen("tcp", ":3000")
	handleError(err)

	go fs.handleSignals()

	for {
		conn, err := event.Accept()
		if err != nil {
			select {
			case <-fs.shutdownSignal:
				return
			default:
				log.Println("Error accepting connection ->", err)
			}
		}

		fs.waitGroup.Add(1)
		go fs.readLoop(conn)
	}
}

func (fs *FileServer) readLoop(conn net.Conn) {
	defer fs.waitGroup.Done()

	var size int64
	err := binary.Read(conn, binary.LittleEndian, &size)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading size ->", err)
		}
		return
	}

	if size <= 0 {
		log.Println("Invalid file size received:", size)
		return
	}

	// Read file data directly from the connection
	fileData := make([]byte, size)
	n, err := io.ReadFull(conn, fileData)
	if err != nil {
		log.Println("Error reading data ->", err)
		return
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("%d bytes of data received over the network\n", n)
	fmt.Println(fileData)
}

func sendFile(size int) error {
	if size <= 0 {
		log.Println("Invalid file size:", size)
		return fmt.Errorf("invalid file size")
	}

	file := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		log.Println("Error generating random data ->", err)
		return err
	}

	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Println("Error connecting to server ->", err)
		return err
	}
	defer conn.Close()

	handleError(binary.Write(conn, binary.LittleEndian, int64(size)))

	n, err := conn.Write(file)
	if err != nil {
		log.Println("Error writing data ->", err)
		return err
	}

	fmt.Printf("%d bytes of data written over the network\n", n)
	return nil
}

func main() {
	fileServer := &FileServer{
		shutdownSignal: make(chan os.Signal, 1),
	}

	go func() {
		time.Sleep(4 * time.Second)
		handleError(sendFile(4000))
	}()

	start(fileServer)

	// Wait for all goroutines to finish before exiting
	fileServer.waitGroup.Wait()
}
