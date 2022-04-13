package test

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "lightning-engine/api/match/v1"
	"testing"
)

var client pb.MatchServiceClient

func TestMain(m *testing.M) {
	conn, _ := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	client = pb.NewMatchServiceClient(conn)
	m.Run()
}

func TestAdd(t *testing.T) {
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

	req = &pb.AddOrderRequest{Order: &pb.Order{
		Id:          "2",
		UserId:      2,
		Pair:        "BTC-USDT",
		Price:       "21000",
		Amount:      "1",
		Side:        "sell",
		Type:        "limit",
		TimeInForce: "GTC",
	}}
	reply, err = client.AddOrder(context.Background(), req)
	fmt.Println(reply, err)
}
func TestCancel(t *testing.T) {
	req := &pb.CancelOrderRequest{
		Pair: "BTC-USDT",
		Id:   "1",
	}
	reply, err := client.CancelOrder(context.Background(), req)
	fmt.Println(reply, err)
}
