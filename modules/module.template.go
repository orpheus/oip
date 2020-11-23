// This is just a template for creating modules.
// See flod and flocore for examples.
// Create a new package named after the node
// you wish to create a module linker for.
// Then copy and paste this into a module.go file.

// This template implements the required interface fns
// needed to be used as a module by the ModuleManager (modules.go)
// Define connections/disconnections and side effects in these functions.

package modules

import (
	"context"
	"github.com/bitspill/flod/rpcclient"
)

const ModuleId = "module_name-default"

var NameModule = &module{
	id: ModuleId,
}

type module struct {
	id     string
	client *rpcclient.Client
	active bool
	config interface{}
	ctx    context.Context
}

// Get the Module ID
func (m *module) ID() string {
	return m.id
}

// Wait for node be found & add rpc client
func (m *module) ConnectToNode(ctx context.Context) {
	// add context to node
	m.ctx = ctx

	//log.Info("Connected", logger.Attrs{"module": m.ID()})

	// set client
	//m.client = client
	// set active
	//m.active = true
}

// Disconnect rpc client
func (m *module) DisconnectNode() {
	// disconnect node
	m.client.Disconnect()
	// set is not active
	m.active = false
	//log.Info("Disconnected", logger.Attrs{"module": m.ID()})
}

// Initialize side effects
func (m *module) Initialize() {}

// Is the rpc connection current
func (m module) Active() bool {
	return m.active
}
