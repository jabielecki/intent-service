/*
 * Copyright 2018 - Juniper Networks
 * Author: Praneet Bachheti
 *
 * main Implementation
 *
 */

package compilation

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tungstenfabric-preview/intent-service/pkg/compilation/config"
	"github.com/tungstenfabric-preview/intent-service/pkg/compilation/watch"
	"github.com/tungstenfabric-preview/intent-service/pkg/compilationif"
	"github.com/tungstenfabric-preview/intent-service/pkg/db/etcd"
	"github.com/tungstenfabric-preview/intent-service/pkg/log"
	"github.com/tungstenfabric-preview/intent-service/pkg/services"
)

// setupService setups all required services and chains them.
func setupService() *compilationif.CompilationService {
	// create services
	compilationService := compilationif.NewCompilationService()

	// chain them
	services.Chain(
		compilationService,
	)

	// return entry service
	return compilationService
}

type locker interface {
	DoWithLock(context.Context, string, time.Duration, func(ctx context.Context) error) error
}

// Store represents data store that is source of events.
type Store interface {
	Create(context.Context, string, []byte) error
	Put(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
	WatchRecursive(context.Context, string, int64) chan etcd.Event
	InTransaction(ctx context.Context, do func(context.Context) error) error
	Close() error
}

//IntentCompilationService represents Intent Compilation Service.
type IntentCompilationService struct {
	config  *config.Config
	Store   Store
	service *compilationif.CompilationService
	locker  locker

	log logrus.FieldLogger
}

// NewIntentCompilationService makes a new Intent Compilation Service
func NewIntentCompilationService() (*IntentCompilationService, error) {
	c := config.ReadConfig()

	e, err := etcd.DialByConfig()
	if err != nil {
		return nil, err
	}

	l, err := etcd.NewDistributedLocker()
	if err != nil {
		return nil, err
	}

	return &IntentCompilationService{
		service: setupService(),
		Store:   etcd.NewClient(e),
		locker:  l,
		config:  &c,
		log:     log.NewLogger(c.DefaultCfg.ServiceName),
	}, nil
}

// handleMessage handles message received from etcd pubsub.
func (ics *IntentCompilationService) handleMessage(
	ctx context.Context, index int64, oper int32, key string, newValue []byte,
) {

	ics.log.Debugf("Index: %d, oper: %d, Got Message %s: %s\n",
		index, oper, key, newValue)

	var skipMessage bool
	if err := ics.Store.InTransaction(ctx, func(ctx context.Context) error {
		skipMessage = true
		storedIndex, err := ics.getStoredIndex(ctx)
		if err != nil {
			ics.log.WithError(err).Debug("Error getting stored message index, skipping the message")
			return nil
		}

		if index <= storedIndex {
			ics.log.Debugf("index %d <= storedIndex %d", index, storedIndex)
			return nil
		}
		ics.log.Debugf("index %d > storedIndex %d", index, storedIndex)

		ics.putStoredIndex(ctx, index)

		skipMessage = false
		return nil
	}); err != nil {
		ics.log.WithError(err).Error("etcd transaction failed")
	}

	if !skipMessage {
		watch.AddJob(ctx, index, oper, key, string(newValue))
		ics.log.Debugf("#goroutines: %d", runtime.NumGoroutine())
	}
}

func (ics *IntentCompilationService) getStoredIndex(ctx context.Context) (int64, error) {
	txn := etcd.GetTxn(ctx)
	messageIndexKey := ics.config.EtcdNotifierCfg.MsgIndexString

	storedIndexData := txn.Get(messageIndexKey)

	storedIndex, err := strconv.ParseInt(string(storedIndexData), 10, 64)
	if err != nil {
		return 0, err
	}

	return storedIndex, nil
}

func (ics *IntentCompilationService) putStoredIndex(ctx context.Context, index int64) {
	txn := etcd.GetTxn(ctx)
	messageIndexKey := ics.config.EtcdNotifierCfg.MsgIndexString

	newIndexStr := strconv.FormatInt(index, 10)
	txn.Put(messageIndexKey, []byte(newIndexStr))
}

// Run runs the IntentCompilationService.
func (ics *IntentCompilationService) Run(ctx context.Context) error {
	ics.log.Debug("Running Service")

	watch.WatcherInit(ics.config.DefaultCfg.MaxJobQueueLen)
	watch.InitDispatcher(ics.config.DefaultCfg.NumberOfWorkers, ics.service.HandleEtcdMessages)

	ics.log.Debug("Setting MessageIndex to 0 (if not exists)")
	err := ics.Store.Create(ctx, ics.config.EtcdNotifierCfg.MsgIndexString, []byte("0"))
	if err != nil {
		ics.log.Println("Cannot Set MessageIndex")
		return err
	}

	// Init watching channel
	watchPath := ics.config.EtcdNotifierCfg.WatchPath
	ics.log.WithField("watchPath", watchPath).Debug("Starting recursive watch")
	eventChan := ics.Store.WatchRecursive(ctx, "/"+watchPath, int64(0))

	watch.RunDispatcher()

	ics.log.Debug("Starting handle loop")
	for {
		select {
		case <-ctx.Done():
			return nil
		case e, ok := <-eventChan:
			if !ok {
				return errors.New("event channel unsuspectingly closed")
			}
			ics.handleMessage(ctx, e.Revision, e.Type, e.Key, e.Value)
		}
	}
}
