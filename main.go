package main

import "crypto/tls"
import "net"
import "fmt"
import "os"
import "os/signal"
import "syscall"

type NewConnectionFunc func(*net.Conn)

type Server struct {
	Config   *tls.Config
	Listener *net.Listener
	Done     chan bool
}

func NewServer() *Server {
	return &Server{
		Config:   nil,
		Listener: nil,
		Done:     make(chan bool, 1),
	}
}

func (s *Server) Initialize() {
	cert, err := tls.LoadX509KeyPair("mykey.pem", "mykey.key")
	if err != nil {
		panic(err)
	}
	s.Config = &tls.Config{
		Certificates: make([]tls.Certificate, 1),
	}
	s.Config.Certificates[0] = cert
	s.Config.BuildNameToCertificate()
}

func (s *Server) Listen(fn NewConnectionFunc) {
	listener, err := tls.Listen("tcp", ":10000", s.Config)
	if err != nil {
		panic(err)
	}
	s.Listener = &listener

	for {
		connection, err := listener.Accept()
		if err != nil {
			select {
			case <-s.Done:
				fmt.Println("Server closed")
			default:
				fmt.Println("Connection error")
			}
			break
		} else {
			fmt.Println(connection.RemoteAddr(), "connected")
			go fn(&connection)
		}
	}
}

func NewConnection(conn *net.Conn) {
	buffer := make([]byte, 2048)
	bytesRead, err := conn.Read(buffer)
	if err != nil || bytesRead <= 0 {
		conn.Close()
		return
	}

	conn.Write([]byte("Echoing back"))
	conn.Write(buffer)
	conn.Close()
}

func main() {

	done := make(chan bool, 1)
	s := NewServer()
	s.Initialize()

	// Capture Ctrl+C signal, and cleanup before terminate
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		s.Done <- true
		done <- true
	}()

	<-done

}
