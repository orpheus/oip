package helpers

import (
	"github.com/spf13/viper"
	"os"
	"runtime/pprof"
)

func CreateCpuProfileFile () {
	oipdCpuProfileFile := viper.GetString("cpuprofile")
	if oipdCpuProfileFile != "" {
		f, profErr := os.Create(oipdCpuProfileFile)

		if profErr != nil {
			log.Error("could not create CPU profile: ", profErr)
		} else {
			defer f.Close()
			if profErr := pprof.StartCPUProfile(f); profErr != nil {
				log.Error("could not start CPU profile: ", profErr)
			}
			defer pprof.StopCPUProfile()
		}
	}

}
