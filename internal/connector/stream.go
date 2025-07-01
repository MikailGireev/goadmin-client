package connector

import (
	"context"
	"log"
	"time"

	pb "github.com/MikailGireev/goadmin-proto/gen/go/proto"
)

type AgentStream struct {
	Stream pb.AdminService_CommunicateClient
	AgentID string
	Ctx context.Context
	Cancel context.CancelFunc
}

func (as *AgentStream) Run() {
	go as.ReadLoop()
	go as.WriteLoop()
}

func (as *AgentStream) ReadLoop() {
	for {
		msg, err := as.Stream.Recv()
		if err != nil {
			log.Printf("read error: %v", err)
			as.Cancel()
			return
		}
		log.Printf("[server] %s", msg.Note)
	}
}

func (as *AgentStream) WriteLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-as.Ctx.Done():
			log.Printf("WriteLoop canceled")
			return
		case <-ticker.C:
			err := as.Stream.Send(&pb.AgentMessage{AgentId: as.AgentID})
			if err != nil {
				log.Printf("send error: %v", err)
				as.Cancel()
				return
			}
		}
	}
}