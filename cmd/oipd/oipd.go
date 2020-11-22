package main

import (
	"context"
	"fmt"
	"github.com/azer/logger"
	"github.com/oipwg/oip/version"
	"github.com/orpheus/oip/modules"
	"github.com/orpheus/oip/modules/flod"
	"github.com/orpheus/oip/util"
)

func main() {
	// create context
	ctx := context.Background()
	ctx, cancelRoot := context.WithCancel(ctx)
	// handle daemon shutdowns
	util.HandleSignalShutdown(cancelRoot)

	// create config files if needed
	util.CreateCpuProfileFile()
	util.CreateMemProfileFile()

	// log version info
	log.Info(" OIP Daemon ", logger.Attrs{
		"commitHash": version.GitCommitHash,
		"buildDate":  version.BuildDate,
		"builtBy":    version.BuiltBy,
		"goVersion":  version.GoVersion,
	})

	// initialize modules
	log.Info("\n Initialize Module Link \n ")
	mm := modules.Initialize(ctx)

	//mm.DeferAllModuleDisconnects()
	//for _, mod := range mm.Modules {
	//	func() {
	//		defer mod.DisconnectNode()
	//	}()
	//}
	fmt.Println(mm.GetModule(flod.ModuleId))

	<-ctx.Done()
	log.Info("Shut down daemon.")
}
