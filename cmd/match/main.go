package main

import (
	"google.golang.org/grpc"
	pb "lightning-engine/api/match/v1"
	"lightning-engine/internal/server"
	"lightning-engine/internal/status"
	"log"
	"net"
)

func main() {
	pairs := []string{"BTC-USDT", "ETH-USDT"}
	app, cleanup, err := wireApp(pairs)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	go app.sysSignalHandle.Begin()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMatchServiceServer(grpcServer, app.server)
	log.Println("[RPC] :8080")
	grpcServer.Serve(lis)
	select {}
}

type app struct {
	status          *status.Status
	server          *server.Server
	sysSignalHandle *status.SysSignalHandle
}

func newApp(st *status.Status, se *server.Server, ss *status.SysSignalHandle) *app {
	return &app{
		status:          st,
		server:          se,
		sysSignalHandle: ss,
	}
}
