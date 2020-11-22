package main

import (
	"github.com/azer/logger"
	"github.com/oipwg/oip/version"
	"github.com/orpheus/oip/helpers"
	"github.com/orpheus/oip/modules"
)

func main() {
	helpers.CreateCpuProfileFile()
	helpers.CreateMemProfileFile()

	log.Info(" OIP Daemon ", logger.Attrs{
		"commitHash": version.GitCommitHash,
		"buildDate":  version.BuildDate,
		"builtBy":    version.BuiltBy,
		"goVersion":  version.GoVersion,
	})

	modules.Initialize()
	//log.Info("\n Beginning Module Link \n ")

	//defer flo.Disconnect()
	//tenMinuteCtx, cancel := context.WithTimeout(rootContext, 10*time.Minute)
	//defer cancel()

	helpers.CancelRoot()
}
