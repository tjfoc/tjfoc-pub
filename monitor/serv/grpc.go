package serv

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	ps "github.com/tjfoc/tjfoc/monitor/protos/monitor"

	"github.com/spf13/viper"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmtls"
	"github.com/tjfoc/gmtls/gmcredentials"
	"google.golang.org/grpc"
)

//新建grpc服务
func newGrpcServer() (*grpc.Server, []grpc.DialOption) {
	//是否使用TLS
	useTLS := viper.GetBool("Rpc.UseTLS")
	if useTLS {
		//从本地加载证书和私钥
		cert, err := gmtls.LoadX509KeyPair(viper.GetString("Rpc.TLSCertPath"), viper.GetString("Rpc.TLSKeyPath"))
		if err != nil {
			fmt.Println("gmtls load key pair err:", err)
		}

		certPool := sm2.NewCertPool()
		//读取ca证书
		cacert, err := ioutil.ReadFile(viper.GetString("Rpc.TLSCaPath"))
		if err != nil {
			//log.Error(err)
			fmt.Println(err)
		}

		//把ca证书加入证书池
		certPool.AppendCertsFromPEM(cacert)
		creds := gmcredentials.NewTLS(&gmtls.Config{
			ServerName:   viper.GetString("Rpc.TLSServerName"),
			Certificates: []gmtls.Certificate{cert},
			RootCAs:      certPool,
		})
		return grpc.NewServer(grpc.Creds(gmcredentials.NewTLS(&gmtls.Config{
				Certificates: []gmtls.Certificate{cert},
				ClientCAs:    certPool,
			}))), []grpc.DialOption{grpc.WithTimeout(time.Duration(viper.GetInt("Rpc.TimeOut")) * time.Millisecond),
				grpc.WithTransportCredentials(creds)}

	} else {
		return grpc.NewServer(), []grpc.DialOption{grpc.WithInsecure()}
	}

}

func (s *Service) GrpcServerRun() {
	server, a := newGrpcServer()
	s.grpcOpts = a
	//注册服务
	ps.RegisterMonitorServer(server, s)
	//监听rpc端口
	if listen, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt("Rpc.Port"))); err != nil {
		fmt.Println(err)
		return
	} else {
	retry:
		server.Serve(listen)
		for {
			listen, err = net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt("Rpc.Port")))
			if err != nil {
				goto retry
			}
			time.Sleep(3 * time.Second)
		}
	}
}
