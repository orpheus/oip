package flodrpc

import (
	"context"
	"github.com/azer/logger"
	"net/rpc"
)

const ModuleId = "flodrpc-default"

var FlodRPCModule = &module{
	id: ModuleId,
}

type module struct {
	id     string
	client *rpc.Client
	active bool
	config interface{}
	ctx    context.Context
}

// Get the Module ID
func (m *module) ID() string {
	return m.id
}

// Wait for node be found & add rpc client
func (m *module) ConnectToNode(ctx context.Context, Ready chan bool) {
	// add context to node
	m.ctx = ctx

	c, err := DialRPCNode()
	if err != nil {
		log.Error("Failed to DIAL flod node via rpc", logger.Attrs{"err": err})
		return
	}
	log.Info("Connected Module", logger.Attrs{"module": m.ID()})

	m.client = c
	m.active = true
	Ready <- true
}

func DialRPCNode () (*rpc.Client, error) {
	// wait for connection
	return rpc.Dial("tcp", "localhost:8334")
}

// Disconnect rpc client
func (m *module) DisconnectNode() {
	if m.Active() {
		// disconnect node
		m.client.Close()
		// set is not active
		m.active = false
		log.Info("Disconnected", logger.Attrs{"module": m.ID()})
	}
}

// Initialize side effects
func (m *module) Initialize() {}

// Is the rpc connection current
func (m module) Active() bool {
	return m.active
}
