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
	"github.com/tjfoc/tjfoc/core/blockchain"
	"github.com/tjfoc/tjfoc/core/transaction"
)

func tCallBack(usrData interface{}, tx *transaction.Transaction) int {
	btxsp, _ := usrData.(*blockchain.BlockTxs)
	btxs := *btxsp
	p, _ := btxs.UsrData.(*peer)
	hashData, _ := tx.Hash(p.cryptPlug)
	//logger.Infof("tryToDecryptTx txid %x\n", hashData)
	smartContract, smartContractArgs, err := tryToDecryptTx(p, tx)
	if err != nil {
		logger.Warningf("tryToDecryptTx txid:%x err:%v,", hashData, err)
		panic(1)
	}
	//txData, _ := tx.Show()
	//logger.Infof("upd tx, txid:%x txLen:%d", hashData, len(txData))
	btxs.SmartContract = append(btxs.SmartContract, smartContract)
	btxs.SmartContractArgs = append(btxs.SmartContractArgs, smartContractArgs)
	btxs.Hashs = append(btxs.Hashs, hashData)
	*btxsp = btxs
	return 0
}

func tryToDecryptTx(p *peer, tx *transaction.Transaction) ([]byte, [][]byte, error) {
	//logger.Infof("in sc.tryToDecryptTx")
	//defer logger.Infof("end sc.tryToDecryptTx")
	smartContract := []byte{}
	smartContractArgs := [][]byte{}
	if !tx.IsPrivacy() {
		//logger.Infof("tx is not privacy")
		return tx.TryToUnmarshalDeal(tx.GetTransactionDeal())
	}

	key := tx.GetCipherKey(p.peerid)
	//logger.Infof("cur %s GetCipherKey Key:%x", p.peerid, key)
	if key == nil || string(key) == "" {
		logger.Infof("this peer ignore the transaction,return")
		return smartContract, smartContractArgs, nil
	}
	if k, err := p.cryptPlug.Decrypt(key); err != nil {
		logger.Warningf("Decrypt Key with privateKey err peer`key:%x  err:%v", key, err)
		return smartContract, smartContractArgs, err
	} else {
		//logger.Infof("decrypt k:%x", k)
		if plaintext, err := p.cryptPlug.DecryptBlock(k, tx.GetTransactionDeal()); err != nil {
			logger.Warningf("Cipher Decrypt txdeal [k=%x] err %v", k, err)
			return smartContract, smartContractArgs, err
		} else {
			return tx.TryToUnmarshalDeal(plaintext)
		}
	}
}
