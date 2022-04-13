package server

import (
	"context"
	"github.com/shopspring/decimal"
	pb "lightning-engine/api/match/v1"
	"lightning-engine/internal/match"
	"lightning-engine/internal/status"
	"lightning-engine/models"
)

type Server struct {
	pb.UnimplementedMatchServiceServer
	pool   *match.MatchPool
	status *status.Status
}

func NewServer(status *status.Status, pool *match.MatchPool) *Server {
	return &Server{
		pool:   pool,
		status: status,
	}
}

func (s *Server) AddOrder(ctx context.Context, in *pb.AddOrderRequest) (*pb.AddOrderReply, error) {
	price, err := decimal.NewFromString(in.Order.Price)
	if err != nil {
		return &pb.AddOrderReply{Result: &pb.ReplyResult{Code: 400, Msg: "price error"}}, err
	}
	amount, err := decimal.NewFromString(in.Order.Amount)
	if err != nil {
		return &pb.AddOrderReply{Result: &pb.ReplyResult{Code: 400, Msg: "amount error"}}, err
	}
	order := &models.Order{
		Id:          in.Order.Id,
		UserId:      in.Order.UserId,
		Pair:        in.Order.Pair,
		Price:       price,
		Amount:      amount,
		Side:        in.Order.Side,
		Type:        in.Order.Type,
		TimeInForce: in.Order.TimeInForce,
	}
	err = s.pool.AddOrder(order)
	if err != nil {
		return &pb.AddOrderReply{Result: &pb.ReplyResult{Code: 400, Msg: err.Error()}}, err
	}
	return &pb.AddOrderReply{Result: &pb.ReplyResult{Code: 0, Msg: "success"}}, nil
}

// CancelOrder 撤单
func (s *Server) CancelOrder(ctx context.Context, in *pb.CancelOrderRequest) (*pb.CancelOrderReply, error) {
	err := s.pool.CancelOrder(in.Pair, in.Id)
	if err != nil {
		return &pb.CancelOrderReply{Result: &pb.ReplyResult{Code: 400, Msg: err.Error()}}, err
	}
	return &pb.CancelOrderReply{Result: &pb.ReplyResult{Code: 0, Msg: "success"}}, nil
}
