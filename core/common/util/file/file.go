package file

import (
	"io/ioutil"
	"os"

	"github.com/tjfoc/tjfoc/core/common/flogging"
)

var logger = flogging.MustGetLogger("file.util")

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func ReadFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
}
