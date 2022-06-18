package misc

import (
	"os"
)

//	=====================================================================
//	function name: IsFile
//	function type: public
//	function receiver: na
//	check path status, return bool
//	=====================================================================
func IsFile(path string) bool {
	if fi, err := os.Stat(path); err != nil || !fi.Mode().IsRegular() {
		return false
	}

	return true
}
