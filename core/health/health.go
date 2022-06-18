package health

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/worldstate"
	"github.com/tjfoc/tjfoc/protos/peer"
	"time"
)

var healthLogger = flogging.MustGetLogger("health")

type healthStatus struct {
	memTotal         uint64
	memUsedPercent   float64
	diskTotal        map[string]uint64
	diskUsedPercent  map[string]float64
	singleCPUPercent []float64
	totalCPUPercent  float64

	wsRateTotal      float64
	startTimeTotal   int64
	wsPushTimesTotal int64

	wsRateFromLast      float64
	startTimeFromLast   int64
	wsPushTimesFromLast int64

	wsRateEverySec10      float64
	wsPushTimesEverySec10 int64
}

func (h *healthStatus) GetPeerStatusInfo() *peer.PeerStatusInfo {
	p := new(peer.PeerStatusInfo)
	p.DiskTotal = h.diskTotal
	p.DiskUsedPercent = h.diskUsedPercent
	p.MemTotal = h.memTotal
	p.MemUsedPercent = h.memUsedPercent
	p.SingleCPUUsedPercent = h.singleCPUPercent
	p.TotalCPUUsedPercent = h.totalCPUPercent

	t := time.Now().Unix()
	h.wsRateFromLast = float64(h.wsPushTimesFromLast) / float64(t-h.startTimeFromLast)
	h.wsRateTotal = float64(h.wsPushTimesTotal) / float64(t-h.startTimeTotal)
	h.wsPushTimesFromLast = 0
	h.startTimeFromLast = t

	p.WsRateTotal = h.wsRateTotal
	p.WsRateFromLast = h.wsRateFromLast
	p.WsRateEverySec10 = h.wsRateEverySec10
	return p
}
func (h *healthStatus) process() {
	tker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-tker.C:
			h.update()
		case <-worldstate.GetWorldState().GetHealthCh():
			h.processWS()
		}
	}
}
func (h *healthStatus) processWS() {
	h.wsPushTimesTotal += 1
	h.wsPushTimesEverySec10 += 1
	h.wsPushTimesFromLast += 1
}
func (h *healthStatus) update() {
	h.wsRateEverySec10 = float64(h.wsPushTimesEverySec10) / float64(10.0)
	h.wsPushTimesEverySec10 = 0

	p, e := cpu.Percent(time.Second, false)
	if e != nil {
		healthLogger.Error("Get cpu percent total error:", e)
	} else {
		h.totalCPUPercent = p[0]
	}
	p, e = cpu.Percent(time.Second, true)
	if e != nil {
		healthLogger.Error("Get cpu percent single error:", e)
	} else {
		for i, v := range p {
			if i > (len(h.singleCPUPercent) - 1) {
				h.singleCPUPercent = append(h.singleCPUPercent, v)
			} else {
				h.singleCPUPercent[i] = v
			}
		}
	}
	m, e := mem.VirtualMemory()
	if e != nil {
		healthLogger.Error("Get mem info error:", e)
	} else {
		h.memTotal = m.Total
		h.memUsedPercent = m.UsedPercent
	}
	ps, e := disk.Partitions(false)
	if e != nil {
		healthLogger.Error("Get Partitions error:", e)
	} else {
		for _, v := range ps {
			u, e := disk.Usage(v.Mountpoint)
			if e != nil {
				healthLogger.Errorf("Get disk info on Mountpoint:%s error:%s\n", v.Mountpoint, e)
			} else {
				h.diskTotal[v.Mountpoint] = u.Total
				h.diskUsedPercent[v.Mountpoint] = u.UsedPercent
			}
		}
	}
}
