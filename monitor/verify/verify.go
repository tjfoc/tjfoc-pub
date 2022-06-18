package verify

import (
	"bytes"

	logging "github.com/tjfoc/tjfoc/core/common/flogging"
)

var logger = logging.MustGetLogger("verify")

func CompareResult(results [][]byte) ([]byte, int) {
	res, num := getTheMost(results)
	if res != nil {
		return res, num
	}
	return nil, 0
}

type resultStatistics struct {
	results     [][]byte
	resultTimes []int
}

//从交易结果的集合中，寻找出现频率最高的，并且覆盖率大于50%的值，并返回
func getTheMost(results [][]byte) ([]byte, int) {
	if results == nil {
		return nil, 0
	}
	r := &resultStatistics{
		results:     [][]byte{},
		resultTimes: []int{}, //resultTimes
	}
	l := len(results)
	logger.Info("接受的交易结果的个数:", l)
	if len(r.results) == 0 {
		r.results = append(r.results, results[0])
		r.resultTimes = append(r.resultTimes, 0)
	}
	for _, v0 := range results {
		ll := len(r.results)
		for k, v1 := range r.results {
			if bytes.Compare(v0, v1) == 0 {
				r.resultTimes[k] += 1
				break
			}
			if k == ll-1 && bytes.Compare(v0, v1) != 0 {
				r.results = append(r.results, v0)
				r.resultTimes = append(r.resultTimes, 1)
			}
		}
	}
	res := 0
	index := 0
	for k, v := range r.resultTimes {
		if v >= res {
			res = v   //v为高频结果出现的次数
			index = k //k为高频结果出现的索引值
		}
	}
	//如果高频结果出现的次数大于50%，返回结果
	if res > l/2 {
		logger.Info("高频结果出现次数大于百分之50！")
		return r.results[index], res
	}

	return nil, 0 //高频结果出现的次数不大于百分之50，交易为失败
}
