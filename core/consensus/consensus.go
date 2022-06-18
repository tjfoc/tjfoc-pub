package consensus

type Consensus interface {
	AppendEntries([]byte)

	IsLeader() bool

	Start()

	LeaderID() string

	Peers() []PeerInfo

	UpdatePeer(typ int, id, addr string) bool
}

type PeerInfo struct {
	ID       string //id
	Addr     string //address
	Model    int    //启动类型 0-核心节点参与共识；1-非核心节点不参与共识
	IsLeader bool   //是否leader
}
