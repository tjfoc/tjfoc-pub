package cmd

import (
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc/stats"
)

//statshandler 需实现4个方法，只用到2个连接相关的方法，TagConn() 和 HandleConn(),
//另外2个 TagRPC() 和 HandleRPC() 用于RPC统计, 实现为空
type statshandler struct{}

// TagConn 用来给连接打个标签，以此来标识连接(实在是找不出还有什么办法来标识连接).
// 这个标签是个指针，可保证每个连接唯一。
// 将该指针添加到上下文中去，键为 connCtxKey{}.
func (h *statshandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, connCtxKey{}, info)
}

// HandleConn 会在连接开始和结束时被调用，分别会输入不同的状态.
func (h *statshandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	tag, ok := getConnTagFromContext(ctx)
	if !ok {
		logger.Fatal("can not get conn tag")
	}

	connsMutex.Lock()
	defer connsMutex.Unlock()

	switch s.(type) {
	case *stats.ConnBegin:
		conns[tag] = ""
		logger.Errorf("begin conn, tag = (%p) remoteAddr:%s, now connections = %d", tag, tag.RemoteAddr, len(conns))
	case *stats.ConnEnd:
		delete(conns, tag)
		logger.Errorf("end conn, tag = (%p) remoteAddr:%s, now connections = %d", tag, tag.RemoteAddr, len(conns))
	default:
		logger.Errorf("illegal ConnStats type")
	}
}

// TagRPC 为空.
func (h *statshandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC 为空.
func (h *statshandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
}

var connsMutex sync.Mutex
var conns map[*stats.ConnTagInfo]string = make(map[*stats.ConnTagInfo]string)

type connCtxKey struct{}

func getConnTagFromContext(ctx context.Context) (*stats.ConnTagInfo, bool) {
	tag, ok := ctx.Value(connCtxKey{}).(*stats.ConnTagInfo)
	return tag, ok
}

// func printInfo() {
// 	for {
// 		time.Sleep(time.Second * 30)
// 		connsMutex.Lock()
// 		logger.Errorf("client connections:%d", len(conns))
// 		connsMutex.Unlock()
// 	}

// }
