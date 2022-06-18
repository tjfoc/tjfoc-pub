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
package config

type PeerConfig struct {
	PackTime  int32
	Typ       string
	Admin     string
	Self      IdConfig
	Log       LogConfig
	Rpc       RpcConfig
	Node      NodeConfig
	Crypt     CryptConfig
	StorePath StoreConfig
	Members   MemberConfig
}

type StoreConfig struct {
	CertStorePath  string
	Path           string
	WorldStatePath string
}

type NodeConfig struct {
	NodeCertPath string
	NodeKeyPath  string
	RootKeyPath  string
	RootCertPath string
}

type LogConfig struct {
	Level     string
	LogServer string
	LogDir    string
	LogFile   string
}

type RpcConfig struct {
	Port          int
	Timeout       int
	UseTLS        bool
	TLSCaPath     string
	TLSKeyPath    string
	TLSCertPath   string
	TLSServerName string
}

type CryptConfig struct {
	KeyTyp   string
	HashTyp  string
	KeyPath  string
	TxVerify bool
}

type IdConfig struct {
	Typ  int
	Id   string
	Addr string
}

type P2pConfig struct {
	Cycle int
}

type MemberConfig struct {
	Port  int
	P2P   P2pConfig
	Peers []IdConfig
}
