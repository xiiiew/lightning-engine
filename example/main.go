package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "lightning-engine/api/match/v1"
	"lightning-engine/cmd/match"
	"log"
	"net"
	"time"
)

func main() {
	// run server
	go server()
	// run client
	clientServer()
}

func server() {

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

func clientServer() {
	conn, _ := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	client := pb.NewMatchServiceClient(conn)

	for {
		sendBuy(client)
		sendSell(client)
		time.Sleep(1 * time.Second)
	}
}

func sendBuy(client pb.MatchServiceClient) {
	req := &pb.AddOrderRequest{Order: &pb.Order{
		Id:          "1",
		UserId:      2,
		Pair:        "BTC-USDT",
		Price:       "21000",
		Amount:      "2",
		Side:        "buy",
		Type:        "limit",
		TimeInForce: "GTC",
	}}
	reply, err := client.AddOrder(context.Background(), req)
	fmt.Println("send buy", reply, err)
}

func sendSell(client pb.MatchServiceClient) {
	req := &pb.AddOrderRequest{Order: &pb.Order{
		Id:          "2",
		UserId:      2,
		Pair:        "BTC-USDT",
		Price:       "21000",
		Amount:      "1",
		Side:        "sell",
		Type:        "limit",
		TimeInForce: "GTC",
	}}
	reply, err := client.AddOrder(context.Background(), req)
	fmt.Println("send sell", reply, err)
}
