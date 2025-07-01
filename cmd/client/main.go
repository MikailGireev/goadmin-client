package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MikailGireev/goadmin-client/internal/connector"
	pb "github.com/MikailGireev/goadmin-proto/gen/go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "gRPC server host:port")
	agentId := flag.String("id", "dev-1", "unique agent indentifier")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func () {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Printf("interrupt signal received, shutting down...")
		cancel()
	}()

	kp := keepalive.ClientParameters{Time: 30 * time.Second}
	conn, err := grpc.Dial(
		*serverAddr,
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(kp),
	)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewAdminServiceClient(conn)
	stream, err := client.Communicate(ctx)
	if err != nil {
		log.Fatalf("communicate: %v", err)
	}

	if err := stream.Send(&pb.AgentMessage{AgentId: *agentId}); err != nil {
		log.Fatalf("send: %v", err)
	}

	log.Printf("agent %s connected to %s", *agentId, *serverAddr)

	agent := &connector.AgentStream{
		Stream: stream,
		AgentID: *agentId,
		Ctx: ctx,
		Cancel: cancel,
	}

	agent.Run()

	<-ctx.Done()
	log.Printf("agent %s disconnected from %s", *agentId, *serverAddr)
}