package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	pb "github.com/MikailGireev/goadmin-proto/gen/go/proto"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "gRPC server host:port")
	agentId := flag.String("id", "dev-1", "unique agent indentifier")
	flag.Parse()

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

	ctx := context.Background()
	client := pb.NewAdminServiceClient(conn)
	stream, err := client.Communicate(ctx)
	if err != nil {
		log.Fatalf("communicate: %v", err)
	}

	if err := stream.Send(&pb.AgentMessage{AgentId: *agentId}); err != nil {
		log.Fatalf("send: %v", err)
	}

	log.Printf("agent %s connected to %s", *agentId, *serverAddr)

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatalf("recv %v", err)
		}
		log.Printf("server note: %q", msg.Note)
	}
}