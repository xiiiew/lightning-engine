//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"lightning-engine/internal/match"
	"lightning-engine/internal/server"
	"lightning-engine/internal/status"
	"lightning-engine/mq"
)

func wireApp(pair []string) (*app, func(), error) {
	panic(wire.Build(match.ProviderSet, server.ProviderSet, status.ProviderSet, mq.ProviderSet, newApp))
}
