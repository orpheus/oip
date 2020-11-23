package modules

import (
	"context"
	"github.com/orpheus/oip/modules/flocore"
)

type moduleMap map[string]Module

// Add modules here until I can find a way to auto add them
var modules = map[string]Module{
	//flod.FlodModule.ID(): flod.FlodModule,
	flocore.FlocoreModule.ID(): flocore.FlocoreModule,
}

type ModuleManager struct {
	Modules moduleMap
}

type Module interface {
	ID() string
	ConnectToNode(ctx context.Context)
	DisconnectNode()
	Active() bool
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

func (m *ModuleManager) DisconnectNodes() {
	for _, mod := range m.Modules {
		mod.DisconnectNode()
	}
}
