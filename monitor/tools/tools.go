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
package tools

import (
	"fmt"
	"hash"

	"github.com/spf13/viper"

	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/tjfoc/core/miscellaneous"

	pt "github.com/tjfoc/tjfoc/monitor/protos/privateTransaction"
)

func ParseAndHashTx(in *pt.Transaction) []byte {
	if in == nil {
		fmt.Println("empty input")
		return nil
	}
	buf := []byte{}
	//交易头
	buf = append(buf, miscellaneous.E32func(in.Header.Version)...)
	buf = append(buf, miscellaneous.E64func(in.Header.Timestamp)...)
	buf = append(buf, miscellaneous.E32func(in.Header.Privacy)...)
	//相关方
	count := uint32(0)
	peers := []byte{}

	for _, v := range in.Peers {
		peers = append(peers, miscellaneous.E32func(uint32(len(v.Id)))...)
		peers = append(peers, v.Id...)

		peers = append(peers, miscellaneous.E32func(uint32(len(v.Key)))...)
		peers = append(peers, v.Key...)
		count++
	}

	buf = append(buf, miscellaneous.E32func(count)...)

	//buff = append(E32func(count), buff...)
	buf = append(buf, peers...)
	//交易数据
	buf = append(buf, miscellaneous.E32func(uint32(len(in.Data)))...)
	buf = append(buf, in.Data...)

	//TODO:哈希算法的选择
	var ha hash.Hash
	switch viper.GetString("Crypto.Hash") {
	case "sm3":
		ha = sm3.New()
	default:
	}
	hashBuf, _ := miscellaneous.GenHash(ha, buf)

	return hashBuf
}
