package main

import (
	"google.golang.org/grpc"
	pb "lightning-engine/api/match/v1"
	"lightning-engine/cmd/match"
	"log"
	"net"
)

func main() {
	pairs := []string{"BTC-USDT", "ETH-USDT"}
	app, cleanup, err := match.WireApp(pairs)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	go app.SysSignalHandle.Begin()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMatchServiceServer(grpcServer, app.Server)
	log.Println("[RPC] :8080")
	grpcServer.Serve(lis)
	select {}
}
