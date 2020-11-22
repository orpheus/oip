package flod

import (
	"github.com/azer/logger"
)

const moduleId = "flod-default"

var FlodModule = &module{
	ID: moduleId,
}

type module struct {
	ID string
	Client interface{}
	active bool
	config interface{}
}

// Get the Module ID
func (m module) GetId () string {
	return m.ID
}
func (m module) ConnectToNode () {
	log.Info("Connected", logger.Attrs{"module": m.GetId()})
}
func (m module) DisconnectNode () {
	log.Info("Disconnected", logger.Attrs{"module": m.GetId()})
}
func (m module) Initialize () {}
func (m module) IsActive () {}
