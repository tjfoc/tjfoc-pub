package check

import (
	"errors"
	"regexp"
	"strings"
)

var words []string

var illegalpack []string
var ischecking bool

const (
	CheckSuccess = iota
	CheckFailed
	CheckBusy
)

// StartCheck 判断智能合约中是否有非法包---合约中不能使用随机数 本机IP等无法共识的元素
func StartCheck(ccSrc []byte) (int, error) {
	words = make([]string, 0)
	if ischecking {
		return CheckBusy, errors.New("check system is busy!")
	}
	ischecking = true
	src := string(ccSrc)
	if findGo(src) {
		ischecking = false
		return CheckFailed, errors.New("Check system find illegal keywords!\nchaincode cann't use goroutine(go keyword)")
	}
	firstKey := strings.Split(src, "func")
	secondKey := strings.Split(firstKey[0], "var")
	thirdKey := strings.Split(secondKey[0], "const")
	fourKey := strings.Split(thirdKey[0], "type")
	for _, w := range illegalpack {
		if ok, _ := regexp.MatchString(w, fourKey[0]); ok {
			ischecking = false
			return CheckFailed, errors.New("Check system find illegal keywords!\nIllegal keywords include 'time','math/rand','crypto','net','os','syscall','runtime'.")
		}
	}
	ischecking = false
	return CheckSuccess, nil
}
func init() {
	illegalpack = make([]string, 0)
	illegalpack = append(illegalpack, "\"math/rand\"")
	illegalpack = append(illegalpack, "\"crypto\"")
	illegalpack = append(illegalpack, "\"net\"")
	illegalpack = append(illegalpack, "\"os\"")
	illegalpack = append(illegalpack, "\"syscall\"")
	illegalpack = append(illegalpack, "\"time\"")
	illegalpack = append(illegalpack, "\"runtime\"")
}

func findGo(src string) bool {
	goKey := strings.SplitN(src, "func", 2)
	ss := goKey[1]
	aa := strings.Contains(ss, "\ngo ")
	bb := strings.Contains(ss, "\tgo ")
	cc := strings.Contains(ss, " go ")
	if aa || bb || cc {
		return true
	}
	return false
}
