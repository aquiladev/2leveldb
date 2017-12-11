package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/aquiladev/2leveldb/source"
	azureStorageTableSource "github.com/aquiladev/2leveldb/source/azure/storage/table"
	"fmt"
)

type converter struct {
	started     int32
	shutdown    int32
	startupTime int64
	quit        chan struct{}
	wg          sync.WaitGroup

	cfg *config
}

var (
	sourceMap = map[string]func(*source.Config) source.ISource{
		"azure.storage.table": azureStorageTableSource.New,
	}
)

func (c *converter) start() {
	sourceConstructor := sourceMap[c.cfg.Source.Type]
	src := sourceConstructor(c.cfg.Source)

	fmt.Println(src)
	c.wg.Done()
}

func (c *converter) Start() {
	// Already started?
	if atomic.AddInt32(&c.started, 1) != 1 {
		return
	}

	converterLog.Trace("Starting worker")

	// Converter startup time. Used for the uptime command for uptime calculation.
	c.startupTime = time.Now().Unix()

	c.wg.Add(1)
	go c.start()
}

func (c *converter) Stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&c.shutdown, 1) != 1 {
		converterLog.Info("Converter is already in the process of shutting down")
		return nil
	}

	converterLog.Warn("Converter shutting down")

	// Signal the remaining goroutines to quit.
	close(c.quit)
	return nil
}

func (c *converter) WaitForShutdown() {
	c.wg.Wait()
}

func newConverter(cfg *config) (*converter, error) {
	return &converter{
		cfg:  cfg,
		quit: make(chan struct{}),
	}, nil
}
