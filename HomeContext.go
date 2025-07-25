// HomeContext
package main

import (
	"os"
	"path"
)

var wHeader = "Default"

func MakeHomeContext(baseDir string) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		baseDir = path.Join(homeDir, "Documents", "FolkTuneWB")
		_, err = os.Stat(baseDir)
		if err != nil {
			err = os.MkdirAll(baseDir, 0777)
			if err != nil {
				panic(err)
			}
		}
	} else {
		wHeader = path.Base(baseDir)
	}
	err := os.Chdir(baseDir)
	if err != nil {
		panic(err)
	}

	os.Mkdir("tmp", 0777)

}
