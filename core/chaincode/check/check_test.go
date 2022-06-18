package check

import "testing"
import "os"
import "io/ioutil"
import "fmt"

func TestStartCheck(t *testing.T) {
	f, _ := os.Open("testfile")
	src, _ := ioutil.ReadAll(f)
	retstatus, err := StartCheck(src)
	if retstatus == CheckSuccess {
		fmt.Println("success")
	} else if retstatus == CheckFailed {
		fmt.Println("failed")
		fmt.Println(err)
	}
}
