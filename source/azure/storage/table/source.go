package table

import (
	"github.com/aquiladev/2leveldb/source"
)

type AzureStorageTableSource struct {
	cfg *source.Config
}

func New(cfg *source.Config) source.ISource {
	return &AzureStorageTableSource{
		cfg: cfg,
	}
}
