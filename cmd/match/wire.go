//go:build wireinject
// +build wireinject

package match

import (
	"github.com/google/wire"
	"lightning-engine/internal/match"
	"lightning-engine/internal/server"
	"lightning-engine/internal/status"
	"lightning-engine/mq"
)

func wireApp(pair []string) (*App, func(), error) {
	panic(wire.Build(match.ProviderSet, server.ProviderSet, status.ProviderSet, mq.ProviderSet, newApp))
}
