package main

import "crypto/tls"
import "net"
import "fmt"
import "os"
import "os/signal"
import "syscall"

var done chan bool

type NewConnectionFunc func(*net.Conn)

type Server struct {
	Config   *tls.Config
	Listener *net.Listener
	Quit     chan bool
}

func NewServer() *Server {
	return &Server{
		Config:   nil,
		Listener: nil,
		Quit:     make(chan bool, 1),
	}
}

func (s *Server) Initialize() {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
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
	fmt.Println("Server listening")

MainLoop:
	for {
		connection, err := listener.Accept()
		if err != nil {
			select {
			case <-s.Quit:
				fmt.Println("Server closed")
			default:
				fmt.Println("Connection error")
			}
			break MainLoop
		} else {
			fmt.Println(connection.RemoteAddr(), "connected")
			go fn(&connection)
		}
	}
	done <- true
}

func (s *Server) Close() {
	s.Quit <- true
	(*s.Listener).Close()
}

func NewConnection(conn *net.Conn) {
	c := *conn
	bytesSent, err := c.Write([]byte("Data from server"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Bytes Sent:", bytesSent)
	c.Close()
}

func main() {

	done = make(chan bool, 1)
	s := NewServer()
	s.Initialize()
	go s.Listen(NewConnection)

	// Capture Ctrl+C signal, and cleanup before terminate
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		s.Close()
	}()

	<-done

}
