package health

import "github.com/tjfoc/tjfoc/protos/peer"
import "time"

type Health_interface interface {
	GetPeerStatusInfo() *peer.PeerStatusInfo
}

var hs *healthStatus

func GetInstance() Health_interface {
	if hs == nil {
		hs := new(healthStatus)
		hs.diskTotal = make(map[string]uint64)
		hs.diskUsedPercent = make(map[string]float64)
		hs.singleCPUPercent = make([]float64, 0)
		hs.wsPushTimesTotal = 0
		hs.wsPushTimesFromLast = 0
		hs.wsPushTimesEverySec10 = 0
		t := time.Now().Unix()
		hs.startTimeTotal = t
		hs.startTimeFromLast = t
		go hs.process()
	}
	return hs
}
