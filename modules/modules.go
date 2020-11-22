package modules

import (
	"context"
	"github.com/orpheus/oip/modules/flod"
)

type moduleMap map[string]Module

// Add modules here until I can find a way to auto add them
var modules = map[string]Module{
	flod.FlodModule.GetId(): flod.FlodModule,
}

type ModuleManager struct {
	Modules moduleMap
}

type Module interface {
	GetId() string
	ConnectToNode(ctx context.Context)
	DisconnectNode()
	IsActive() bool
	Initialize()
}

func Initialize(ctx context.Context) *ModuleManager {
	mm := &ModuleManager{
		Modules: modules,
	}
	for _, mod := range modules {
		mod.ConnectToNode(ctx)
	}
	return mm
}

func (m *ModuleManager) GetModule(id string) Module {
	return m.Modules[id]
}

func (m *ModuleManager) DeferAllModuleDisconnects() {
	for _, mod := range m.Modules {
		func() {
			defer mod.DisconnectNode()
		}()
	}
}
