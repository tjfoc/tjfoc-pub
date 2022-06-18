/*
Copyright Suzhou Tongji Fintech Research Institute 2018 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/transaction"
	"github.com/tjfoc/tjfoc/core/worldstate"
	pbc "github.com/tjfoc/tjfoc/protos/block"
	ppr "github.com/tjfoc/tjfoc/protos/peer"
	ptx "github.com/tjfoc/tjfoc/protos/transaction"
	"golang.org/x/net/context"
)

func (p *peer) Search(ctx context.Context, in *ppr.SearchMes) (*ppr.SearchRes, error) {
	tempres := make(map[string]string, 0)
	switch in.Type {
	//单key查询
	case 0:
		{
			res := worldstate.GetWorldState().Search(string(in.Key[0]))
			if res.Err == nil {
				tempres[res.Key] = res.Value
			}
		}
	//多key查询
	case 1:
		{
			tempkey := make([]string, 0)
			for _, v := range in.Key {
				tempkey = append(tempkey, string(v))
			}
			res := worldstate.GetWorldState().Searchn(tempkey)
			for _, v := range res {
				if v.Err == nil {
					tempres[v.Key] = v.Value
				}
			}
		}
	//前缀查询
	case 2:
		{
			res, err := worldstate.GetWorldState().SearchPrefix(string(in.Key[0]))
			if err == nil {
				tempres = res
			}
		}
	}
	return &ppr.SearchRes{
		Res: tempres,
	}, nil
}

func (p *peer) NewTransaction(ctx context.Context, in *ptx.Transaction) (*ppr.BlockchainBool, error) {
	// if cmd.Config.Crypt.TxVerify {
	// 	if len(in.Header.BigNum) != 64 {
	// 		return nil, errors.New("error big num!")
	// 	}
	// 	x := in.Header.BigNum[0:32]
	// 	y := in.Header.BigNum[32:]
	// 	n1 := big.NewInt(0)
	// 	n2 := big.NewInt(0)
	// 	ok := verify.Verify(n1.SetBytes(x), n2.SetBytes(y), in.Header.TransactionHash, in.Header.TransactionSign)
	// 	if !ok {
	// 		return nil, errors.New("failed to verify signature!")
	// 	}
	// }
	tx := transaction.New(in.SmartContract, in.SmartContractArgs, in.Header.Version, in.Header.Timestamp)
	if hashData, err := tx.Hash(p.cryptPlug); err != nil || bytes.Compare(hashData, in.Header.TransactionHash) != 0 {
		return &ppr.BlockchainBool{
			Ok:  false,
			Err: fmt.Sprintf("error:%s", err),
		}, err
	} else {
		tx.AddSign(in.Header.TransactionSign)
	}
	if err := p.blockChain.TransactionAdd(tx); err != nil {
		return &ppr.BlockchainBool{
			Ok:  false,
			Err: fmt.Sprintf("error:%s", err),
		}, err
	}
	return &ppr.BlockchainBool{
		Ok: true,
	}, nil
}

func (p *peer) BlockchainGetHeight(ctx context.Context, in *ppr.BlockchainBool) (*ppr.BlockchainNumber, error) {
	return &ppr.BlockchainNumber{
		Number: p.blockChain.Height(),
	}, nil
}

func (p *peer) BlockchainGetBlockByHash(ctx context.Context, in *ppr.BlockchainHash) (*pbc.Block, error) {
	if b := p.blockChain.GetBlockByHash(in.HashData); b == nil {
		return nil, errors.New("BlockchainGetBlockByHash: failed to find block")
	} else {
		buf := [][]byte{}
		txs := b.TransactionList()
		for _, v := range *txs {
			if tmp, err := v.Hash(p.cryptPlug); err == nil {
				buf = append(buf, tmp)
			}
		}
		return &pbc.Block{
			Header: &pbc.BlockHeader{
				Height:          b.Height(),
				Version:         b.Version(),
				BlockHash:       in.HashData,
				Timestamp:       b.Timestamp(),
				PreviousHash:    b.PreviousBlock(),
				WorldStateRoot:  b.WorldStateRoot(),
				TransactionRoot: b.TransactionsRoot(),
			},
			Txs: buf,
		}, nil
	}
}

func (p *peer) BlockchainGetBlockByHeight(ctx context.Context, in *ppr.BlockchainNumber) (*pbc.Block, error) {
	if b := p.blockChain.GetBlockByHeight(in.Number); b == nil {
		return nil, errors.New("BlockchainGetBlockByHeight: failed to find block")
	} else if hashData, err := b.Hash(p.cryptPlug); err != nil {
		return nil, err
	} else {
		buf := [][]byte{}
		txs := b.TransactionList()
		for _, v := range *txs {
			if tmp, err := v.Hash(p.cryptPlug); err == nil {
				buf = append(buf, tmp)
			}
		}
		return &pbc.Block{
			Header: &pbc.BlockHeader{
				Version:         b.Version(),
				Height:          b.Height(),
				Timestamp:       b.Timestamp(),
				BlockHash:       hashData,
				PreviousHash:    b.PreviousBlock(),
				WorldStateRoot:  b.WorldStateRoot(),
				TransactionRoot: b.TransactionsRoot(),
			},
			Txs: buf,
		}, nil
	}
}

func (p *peer) BlockchainGetTransaction(ctx context.Context, in *ppr.BlockchainHash) (*ptx.Transaction, error) {
	if tx := p.blockChain.GetTransaction(in.HashData); tx == nil {
		return nil, errors.New("BlockchainGetTransaction: failed to find transaction")
	} else {
		return &ptx.Transaction{
			Records:           tx.Records(),
			SmartContract:     tx.SmartContract(),
			SmartContractArgs: tx.SmartContractArgs(),
			Header: &ptx.TransactionHeader{
				TransactionHash: in.HashData,
				Version:         tx.Version(),
				Timestamp:       tx.Timestamp(),
				TransactionSign: tx.SignData(),
			},
		}, nil
	}
}

func (p *peer) BlockchainGetTransactionIndex(ctx context.Context, in *ppr.BlockchainHash) (*ppr.BlockchainNumber, error) {
	if a, err := p.blockChain.GetTransactionIndex(in.HashData); err != nil {
		return nil, err
	} else {
		return &ppr.BlockchainNumber{
			Number: uint64(a),
		}, nil
	}
}

func (p *peer) BlockchainGetTransactionBlock(ctx context.Context, in *ppr.BlockchainHash) (*ppr.BlockchainNumber, error) {
	if a, err := p.blockChain.GetTransactionBlock(in.HashData); err != nil {
		return nil, err
	} else {
		return &ppr.BlockchainNumber{
			Number: uint64(a),
		}, nil
	}
}

func (p *peer) GetMemberList(ctx context.Context, in *ppr.BlockchainBool) (*ppr.MemberListInfo, error) {
	mlist := new(ppr.MemberListInfo)
	mlist.MemberList = []*ppr.PeerInfo{}
	peers := p.consensusAPI.Peers()
	for _, v := range peers {
		state := 0
		if v.Model == 1 {
			state = 2
		}
		if v.IsLeader {
			state = 1
		}
		v.Addr = strings.Split(v.Addr, ":")[0] + ":" + strconv.Itoa(Config.Rpc.Port)
		mlist.MemberList = append(mlist.MemberList, &ppr.PeerInfo{
			Id:    string(miscellaneous.Dup([]byte(v.ID))),
			Addr:  string(miscellaneous.Dup([]byte(v.Addr))),
			State: int32(state),
		})
	}
	return mlist, nil
}

func (p *peer) UpdatePeer(ctx context.Context, in *ppr.PeerUpdateInfo) (*ppr.BlockchainBool, error) {
	var adminSign AdminSignature

	if in.Admin != Config.Admin {
		return &ppr.BlockchainBool{
			Ok:  false,
			Err: "Permission prohibition",
		}, nil
	}
	sign := base58.Decode(in.Sign)
	data := base58.Decode(in.Data)
	pubKeyData := base58.Decode(in.Admin)
	_, err := asn1.Unmarshal(sign, &adminSign)
	if err != nil {
		return &ppr.BlockchainBool{
			Ok:  false,
			Err: "Permission prohibition",
		}, nil
	}
	switch {
	case strings.Compare(Config.Crypt.KeyTyp, "sm2") == 0:
		var pubKey sm2.PublicKey

		pubKey.X = new(big.Int)
		pubKey.Y = new(big.Int)
		pubKey.Curve = sm2.P256Sm2()
		pubKey.X.SetBytes(pubKeyData[:32])
		pubKey.Y.SetBytes(pubKeyData[32:64])
		fmt.Printf("x = %x\ny = %x\n", pubKey.X.Bytes(), pubKey.Y.Bytes())
		if ok := sm2.Verify(&pubKey, data, adminSign.R, adminSign.S); !ok {
			return &ppr.BlockchainBool{
				Ok:  false,
				Err: "Permission prohibition",
			}, nil
		}
	case strings.Compare(Config.Crypt.KeyTyp, "ecc") == 0:
		var pubKey ecdsa.PublicKey

		pubKey.X = new(big.Int)
		pubKey.Y = new(big.Int)
		pubKey.Curve = elliptic.P256()
		pubKey.X.SetBytes(pubKeyData[:32])
		pubKey.Y.SetBytes(pubKeyData[32:64])
		if ok := ecdsa.Verify(&pubKey, data, adminSign.R, adminSign.S); !ok {
			return &ppr.BlockchainBool{
				Ok:  false,
				Err: "Permission prohibition",
			}, nil
		}
	default:
		return &ppr.BlockchainBool{
			Ok:  false,
			Err: "Permission prohibition",
		}, nil
	}
	ok := p.consensusAPI.UpdatePeer(int(in.Typ), in.Id, in.Addr)
	return &ppr.BlockchainBool{
		Ok: ok,
	}, nil
}
