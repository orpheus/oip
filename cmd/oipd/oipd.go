package main

import (
	"context"
	"github.com/azer/logger"
	"github.com/oipwg/oip/version"
	"github.com/orpheus/oip/modules"
	"github.com/orpheus/oip/util"
)

func main() {
	// create context
	ctx := context.Background()
	ctx, cancelRoot := context.WithCancel(ctx)

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
	modules.Initialize(ctx)

	// handle daemon shutdowns
	util.CancelRoot(ctx, cancelRoot)
}
