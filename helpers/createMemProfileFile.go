package helpers

import (
	"github.com/spf13/viper"
	"os"
	"runtime/pprof"
)

func CreateMemProfileFile () {
	oipdMemProfileFile := viper.GetString("memprofile")
	if oipdMemProfileFile != "" {
		f, memErr := os.Create(oipdMemProfileFile)
		if memErr != nil {
			log.Error("could not create memory profile: ", memErr)
		} else {
			defer f.Close()
			if memErr := pprof.WriteHeapProfile(f); memErr != nil {
				log.Error("could not start memory profile: ", memErr)
			}
		}
	}
}
