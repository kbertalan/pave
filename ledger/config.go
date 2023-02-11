package ledger

import (
	"encore.dev/config"

	"encore.app/ledger/tb"
)

type TigerBeetleConfig struct {
	ClusterID      config.Uint32
	Addresses      config.Values[string]
	MaxConcurrency config.Uint
}

func (tbc TigerBeetleConfig) NewFactory() *tb.Factory {
	return tb.NewFactory(tb.Config{
		ClusterID:      tbc.ClusterID(),
		Addresses:      tbc.Addresses(),
		MaxConcurrency: tbc.MaxConcurrency(),
	})
}

type ServiceConfig struct {
	TigerBeetle TigerBeetleConfig
}

var cfg = config.Load[*ServiceConfig]()
