package modules

import (
	"context"
	"github.com/azer/logger"
	"github.com/orpheus/oip/modules/flod"
	"github.com/orpheus/oip/modules/flodrpc"
)

type moduleMap map[string]Module

// Add modules here until I can find a way to auto add them
var modules = map[string]Module{
	flod.FlodModule.ID(): flod.FlodModule,
	flodrpc.FlodRPCModule.ID(): flodrpc.FlodRPCModule,
	//flocore.FlocoreModule.ID(): flocore.FlocoreModule,
}

type ModuleManager struct {
	Modules moduleMap
}

type Module interface {
	ID() string
	ConnectToNode(ctx context.Context, Ready chan bool)
	DisconnectNode()
	Active() bool
	Initialize()
}

func Initialize(ctx context.Context) *ModuleManager {
	mm := &ModuleManager{
		Modules: modules,
	}
	modulesReady := make(chan bool, len(modules))

	// concurrently connect to nodes
	for _, mod := range modules {
		go mod.ConnectToNode(ctx, modulesReady)
	}

	// wait for modules to be initialized
	for _, mod := range modules {
		<- modulesReady
		log.Info("Link-Module Initialized", logger.Attrs{"ID": mod.ID()})
	}

	close(modulesReady)
	return mm
}

func (m *ModuleManager) GetModule(id string) Module {
	return m.Modules[id]
}

func (m *ModuleManager) DisconnectNodes() {
	for _, mod := range m.Modules {
		go mod.DisconnectNode()
	}
}
