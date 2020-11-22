package modules

import "github.com/orpheus/oip/modules/flod"

type moduleMap map[string]Module

// Add modules here until I can find a way to auto add them
var modules = map[string]Module{
	flod.FlodModule.GetId(): flod.FlodModule,
}

type ModuleManager struct {
	modules moduleMap
}

type Module interface {
	GetId() string
	ConnectToNode()
	DisconnectNode()
	Initialize()
	IsActive()
}

func Initialize () *ModuleManager {
	mm := &ModuleManager{
		modules: modules,
	}
	for _, mod := range modules {
		mod.ConnectToNode()
	}
	return mm
}
