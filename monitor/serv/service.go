package serv

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	rnd "math/rand"

	//	ps "github.com/tjfoc/tjfoc/monitor/protos/privateTransacion"

	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/tjfoc/gmsm/sm2"
	logging "github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/he"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/store"
	"github.com/tjfoc/tjfoc/monitor/crypto"
	mt "github.com/tjfoc/tjfoc/monitor/protos/monitor"
	pt "github.com/tjfoc/tjfoc/monitor/protos/privateTransaction"

	"github.com/tjfoc/tjfoc/monitor/tools"
	"github.com/tjfoc/tjfoc/monitor/verify"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var logger = logging.MustGetLogger("server")

type Service struct {
	sync.RWMutex
	grpcOpts  []grpc.DialOption
	store     store.Store
	txResults map[string]map[int]*mt.TransactionResults
	timeout   getTimeout
	ran       *big.Int
	priv      *sm2.PrivateKey
}

type getTimeout struct {
	time time.Time
	txId []byte
}

type TX struct {
	Key  []byte
	Part [][]byte
}

func NewService() (*Service, error) {
	db, err := store.NewDb("server.db")
	if err != nil {
		logger.Error("New db error:", err)
		return nil, err
	}

	//rs := make(map[int]ps.TransactionResults)
	s := &Service{
		store:     db,
		txResults: make(map[string]map[int]*mt.TransactionResults),
	}

	s.priv, _ = sm2.ReadPrivateKeyFromPem("./pubkey/monitorKey.pem", nil)
	s.ran, _ = he.RandFieldElement(s.priv.Curve, rand.Reader)

	return s, nil
}

// leveldb中保存的k-v
// 1-txid - struct{Key []byte,Part [][]byte} 其中key为对称密钥 part为相关方集合
// 2-"world_state_enkey"+ key(world中key) - 加密的value(world 中的value)
// 3-"result:"+ key(txid) - 交易结果(加密)
// 4-"world_state_key" + key 加密该value的对称密钥

func (s *Service) NewTransaction(ctx context.Context, in *pt.Transaction) (*pt.Transaction, error) {
	logger.Info("********enter new tx********")
	if in == nil {
		return &pt.Transaction{}, errors.New("Trading data were empty!")
	}
	//验证交易类型
	//1-隐私交易
	if in.Header.Privacy != 1 {
		logger.Error("error tx type")
		return nil, errors.New("error tx type!")
	}

	//对称加密
	key, eData, err := crypto.Encrypt(in.Data)
	if err != nil {
		logger.Error("encryption err:", err)
		return nil, err
	}
	logger.Debugf("交易data明文:%x", in.Data)
	logger.Debugf("######对称密钥:%x", key)
	tx := &TX{
		Key: key,
	}

	logger.Info("相关方个数:", len(in.Peers))
	for k, v := range in.Peers {
		logger.Debugf("peer id:%x", v.Id)
		tx.Part = append(tx.Part, v.Id)
		s := GetPubkeyIndex(v.Id)
		ss := viper.GetString(s)
		pub, err := sm2.ReadPublicKeyFromPem(ss, nil)
		if err != nil {
			logger.Error("read pubkey pem err:", err)
			return nil, err
		}
		//用相关方的公钥加密对称密钥
		enKey, err := pub.Encrypt(key)
		if err != nil {
			logger.Error("pubkey encrypt err:", err)
			return nil, err
		}

		logger.Debugf("######对称密钥密文:%x", enKey)
		in.Peers[k].Key = enKey

		//TODO:为了方便测试做了私钥解密的操作，后期需要删除
		strs := "./pubkey/" + strings.TrimPrefix(ss, "./pubkey/pub")
		logger.Debug("==========address of privatekey===========:", strs)
		priv, err := sm2.ReadPrivateKeyFromPem(strs, nil)
		if err != nil {
			logger.Error(err)
		}

		b, err := priv.Decrypt(enKey)
		if err != nil {
			logger.Error("私钥解密失败:", err)
		}
		logger.Debugf("#####解密后的对称密钥:%x", b)
	}

	buf, err := json.Marshal(tx)
	if err != nil {
		logger.Error("NewTx:json marshal err:", err)
		return nil, err
	}
	//保存交易对应的相关方，以及交易密钥

	//返回加密后的交易数据
	in.Data = eData
	//计算哈希
	in.Txid = tools.ParseAndHashTx(in)
	logger.Debugf("######交易ID:%x", in.Txid)
	err = s.store.Set(in.Txid, buf)
	if err != nil {
		logger.Error("NewTx:store err:", err)
		return nil, err
	}
	return in, nil
}

func (s *Service) QueryKey(ctx context.Context, in *mt.TransactionKey) (*mt.TransactionKey, error) {
	r := rnd.New(rnd.NewSource(time.Now().UnixNano()))
	nonce := r.Int31n(1000000)
	buf := miscellaneous.E32func(uint32(nonce))
	//返回随机数
	in.RandNum = buf
	return in, nil
}

func GetPubkeyIndex(b []byte) string {
	if bytes.Compare(b, []byte("155")) == 0 {
		return "Crypto.APubkeyPath"
	} else if bytes.Compare(b, []byte("156")) == 0 {
		return "Crypto.BPubkeyPath"
	} else if bytes.Compare(b, []byte("157")) == 0 {
		return "Crypto.CPubkeyPath"
	} else if bytes.Compare(b, []byte("158")) == 0 {
		return "Crypto.DPubkeyPath"
	} else if bytes.Compare(b, []byte("159")) == 0 {
		return "Crypto.EPubkeyPath"
	} else {
		return ""
	}
}

func (s *Service) Authentication(ctx context.Context, in *mt.TransactionKey) (*mt.TransactionKey, error) {
	logger.Debug("********enter Authentication********")
	if in == nil {
		logger.Error("Authentication:empty input")
		return nil, errors.New("Authentication:empty input")
	}
	switch viper.GetString("Crypto.Key") {
	case "sm2":
		//TODO:后期公钥是要去数据库中取
		ss := GetPubkeyIndex(in.PeerID)
		pub, err := sm2.ReadPublicKeyFromPem(viper.GetString(ss), nil)
		if err != nil {
			logger.Error("read sm2 pubkey error", err)
			return nil, errors.New("read sm2 pubkey err!")
		}
		//随机数签名的验签
		logger.Infof("随机数:%x\n", in.RandNum)
		logger.Infof("随机数签名:%x\n", in.Sign)
		ok := pub.Verify(in.RandNum, in.Sign)
		if !ok {
			logger.Error("the authentication failed ")
			return nil, errors.New("The authentication failed")
		}
		//获取对称密钥
		sKey, err := s.store.Get(append([]byte("world_state_key"), in.Key...))
		if err != nil {
			return nil, errors.New("not found value of the key")
		}
		logger.Debugf("加密key:%x", sKey)
		in.Value, err = crypto.Sm4Decryption(sKey, in.Secretvalue)
		if err != nil {
			logger.Error("sm4 decryption err:", err)
			return nil, err
		}
		logger.Debugf("===send===value:%x", in.Value)
		if err != nil {
			logger.Error("pubkey encrypt err:", err)
			return nil, err
		}

		i, err := strconv.ParseInt(string(in.Value), 10, 64)
		if err != nil {
			logger.Error("strconv---:", err)
			return nil, err
		}
		logger.Debug("+++++++++value++++++++++++:", i)

		v, err := he.Encrypt(s.priv, new(big.Int).SetBytes(in.Value), s.ran)
		if err != nil {
			logger.Error("he encrypt err:", err)
			return nil, err
		}
		in.Value = v

		logger.Debug("&&&&&&&&&&&&&:", in.Value)
		return in, nil
		//TODO:支持ecdsa
	default:
	}
	logger.Error("error crypto key type")
	return nil, errors.New("error crypto key type")
}

//var ta = []byte{0x04, 0x03, 0x02, 0x01}
func (s *Service) SendTxResult(ctx context.Context, in *mt.TransactionResults) (*mt.Result, error) {
	logger.Debug("+++++enter send result+++++")
	if in == nil {
		logger.Error("send tx result:empty tx result")
		return nil, errors.New("get tx result:empty tx result")
	}
	s.Lock()
	defer s.Unlock()

	m := s.txResults[string(in.TxId)]
	if m != nil {
		length := len(m)
		s.txResults[string(in.TxId)][length] = in
	} else {
		s.txResults[string(in.TxId)] = make(map[int]*mt.TransactionResults)
		s.txResults[string(in.TxId)][0] = in
	}
	return &mt.Result{
		Req: true,
	}, nil
}

type txResult struct {
	results []*mt.TransactionResults
}

func (s *Service) GetTxResult(ctx context.Context, in *mt.TransactionResults) (*mt.TransactionResults, error) {
	if in == nil {
		logger.Error("getTxResult:empty input")
	}
	b, _ := s.store.Get(append([]byte("result:"), in.TxId...))
	if b == nil {
		if bytes.Compare(in.TxId, s.timeout.txId) != 0 {
			s.timeout.txId = in.TxId
			s.timeout.time = time.Now()
		} else {

			outtime := s.timeout.time.Add(time.Second * 3)
			now := time.Now()
			if now.After(outtime) {
				logger.Debug("超时!")
				res := &mt.TransactionResults{}
				b, err := json.Marshal(res)
				if err != nil {
					logger.Error("get tx result marshal failed 11")
					return nil, errors.New("get tx result :marshal failed 11")
				}
				//保存该交易id的空结果
				err = s.store.Set(append([]byte("result:"), in.TxId...), b)
				if err != nil {
					logger.Error(err)
					return nil, errors.New("get tx result:set db value err")
				}
			}
		}
		logger.Errorf("getTxResult error:[not found tx result!]==ID:%x", in.TxId)
		return nil, errors.New("not found tx result")
	}
	//TODO:invalid character 'e' looking for beginning of value
	result := &mt.TransactionResults{}
	err := json.Unmarshal(b, result)
	if err != nil {
		logger.Error("GetTxResult:unmarshal result failed:", err)
		return nil, errors.New("unmarshal result failed")
	}
	return result, nil
}

func (s *Service) SendTotalTxResult(ctx context.Context, in *mt.DealResults) (*mt.DealResults, error) {
	logger.Debug("+++++enter send total result+++++")
	if in == nil {
		return nil, errors.New("empty input")
	}
	res := &mt.DealResults{
		Kv: make(map[string][]byte),
	}
	for key, va := range in.Kv {
		//获取对称密钥
		enKey, _ := s.store.Get(append([]byte("world_state_key"), []byte(key)...))
		if enKey == nil {
			logger.Error("get world state key err")
			return nil, errors.New("not found the enkey")
		}

		v := new(big.Int)

		if in.Tongtai[key] {
			var err error
			logger.Error("encrypt value:", va)
			v, err = he.Decrypt(s.priv, va)
			logger.Error("value:", v.Int64())
			if err != nil {
				logger.Error("decrypt err:", err)
			}

			vaa, err := crypto.Sm4Encryption(enKey, v.Bytes())
			if err != nil {
				logger.Error("sm4 encryption err")
				return nil, errors.New("encrypto error")
			}
			res.Kv[key] = vaa

		} else {
			vaa, err := crypto.Sm4Encryption(enKey, va)
			if err != nil {
				logger.Error("sm4 encryption err")
				return nil, errors.New("encrypto error")
			}
			res.Kv[key] = vaa
		}

	}
	return res, nil
}

func (s *Service) RequireCrypt(ctx context.Context, in *mt.CryptNumber) (*mt.CryptNumber, error) {
	logger.Debug("======enter requireCrypt=====")
	if in == nil {
		logger.Error("empty input")
		return nil, errors.New("empty input")
	}

	v, err := he.Encrypt(s.priv, new(big.Int).SetInt64(in.Value), s.ran)
	if err != nil {
		logger.Error("he encrypt err:", err)
		return nil, err
	}
	in.Svalue = v
	logger.Debug("返回同态加密结果:", v)

	//TODO:

	return in, nil
}

func (s *Service) RequireCompare(ctx context.Context, in *mt.CompareNumber) (*mt.CompareNumber, error) {
	logger.Debug("=====enter requireCompare======")

	if in == nil {
		logger.Error("empty input")
		return nil, errors.New("empty input")
	}

	v, err := he.Decrypt(s.priv, in.Svalue)
	if err != nil {
		logger.Error("decrypt err:", err)
	}

	in.Result = int64(v.Cmp(new(big.Int).SetInt64(in.Value)))
	logger.Debug("++++++compare result++++++", in.Result)
	return in, nil
}

type handleTimeout struct {
	timeout bool
	txId    []byte
}

func (s *Service) HandleTxResults() {
	s.Lock()
	defer s.Unlock()
	var results = [][]byte{}
	//k为string（txid）
	//v 为 map[int] result 某笔交易的结果集合
	for k, v := range s.txResults {
		//logger.Info("获取的交易结果key:", append([]byte("result:"), []byte(k)...))
		buf, _ := s.store.Get(append([]byte("result:"), []byte(k)...))
		if buf != nil {
			logger.Info("交易结果已存在！")
			//delete()
			delete(s.txResults, k)
			break
		}

		buf, _ = s.store.Get([]byte(k))
		if buf == nil {
			logger.Error("交易不存在!")
			return
		}
		tx := &TX{}
		err := json.Unmarshal(buf, tx)
		if err != nil {
			logger.Error("json unmarshal failed:", err)
			return
		}
		//相关方的数量
		num := len(tx.Part)
		//如果交易结果数和相关方相等
		if len(v) == num {
			for _, vv := range v {
				buf, err := json.Marshal(vv)
				if err != nil {
					logger.Error("json marshal failed 1:", err)
					return
				}
				results = append(results, buf)
			}
			buf, _ := verify.CompareResult(results)
			if buf != nil {
				s.store.Set(append([]byte("result:"), []byte(k)...), buf)
				rs := &mt.TransactionResults{}
				err := json.Unmarshal(buf, rs)
				if err != nil {
					logger.Error("json unmarshal failed 2:", err)
					return
				}
				for wk, _ := range rs.Kv {
					//保存对称密钥
					s.store.Set(append([]byte("world_state_key"), []byte(wk)...), tx.Key)
				}
				delete(s.txResults, k)
			}
		}
		if len(v) > num/2 {
			for _, vv := range v {
				buf, err := json.Marshal(vv)
				if err != nil {
					logger.Error("json marshal failed 3")
					return
				}
				results = append(results, buf)
			}
			buf, n := verify.CompareResult(results)
			if n == len(v) && buf != nil {
				s.store.Set(append([]byte("result:"), []byte(k)...), buf)
				rs := &mt.TransactionResults{}
				err := json.Unmarshal(buf, rs)
				if err != nil {
					logger.Error("json unmarshal")
					return
				}
				for wk, _ := range rs.Kv {
					logger.Info("++++++++++++++++++保存:", tx.Key)
					s.store.Set(append([]byte("world_state_key"), []byte(wk)...), tx.Key)
				}
				delete(s.txResults, k)
			}
		}
	}
}

func init() {

}
