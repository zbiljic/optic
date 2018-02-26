package config

import (
	"errors"
	"sync"
)

const maxBuildAttempts = 4

var errPluginReferenceNotFound = errors.New("plugin reference not found")

type plugin struct {
	pluginType    string
	name          string
	config        map[string]interface{}
	buildAttempts int8
}

func (p *plugin) buildAttemptsLimitReached() bool {
	return p.buildAttempts > maxBuildAttempts
}

type pluginBuilder struct {
	// Channel used for building plugins.
	//
	// Since we are iterating though a map, order is not guarantied, so one plugin
	// may reference another before it has been created. If this happens, we just
	// send that plugin again in this channel in the hope that reference will be
	// resolved in the next pass.
	ch    chan *plugin
	endCh chan struct{}
	wg    *sync.WaitGroup
	err   error
}

func newPluginBuilder() *pluginBuilder {
	return &pluginBuilder{
		ch:    make(chan *plugin, 100),
		endCh: make(chan struct{}),
		wg:    new(sync.WaitGroup),
		err:   nil,
	}
}

func (pb *pluginBuilder) addPlugin(p *plugin) {
	pb.wg.Add(1)
	pb.ch <- p
}

func (pb *pluginBuilder) stop() {
	close(pb.ch)
	close(pb.endCh)
}
