// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ca

import (
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmtls"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	util "github.com/tjfoc/tjfoc/core/common/util/file"
	"github.com/tjfoc/tjfoc/core/store"
	"github.com/tjfoc/tjfoc/peer/cmd"
)

var logger = flogging.MustGetLogger("CA")

type cStore struct {
	db store.Store
}

var ips []string
var lock sync.Mutex

var certStore cStore

func init() {
	db, err := store.NewDb(cmd.Config.StorePath.CertStorePath)
	if err != nil {
		logger.Error(err)
		return
	}

	certStore.db = db
}

func getVerifyOptions() (*sm2.VerifyOptions, error) {

	if util.FileExists(cmd.Config.Node.NodeCertPath) {
		cfg, err := util.ReadFile(cmd.Config.Node.RootCertPath)
		if err != nil {
			logger.Fatal("Fail to read root cert!  err === ", err)
		}

		block, rest := pem.Decode(cfg)

		if block == nil {
			logger.Error("No root certificate was found")
			return nil, errors.New("No root certificate was found")
		}
		rootCert, err := sm2.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse root certificate: %s", err)
		}
		rootPool := sm2.NewCertPool()
		rootPool.AddCert(rootCert)
		var intPool *sm2.CertPool
		if len(rest) > 0 {
			intPool = sm2.NewCertPool()
			if !intPool.AppendCertsFromPEM(rest) {
				return nil, errors.New("Failed to add intermediate PEM certificates")
			}
		}
		return &sm2.VerifyOptions{
			Roots:         rootPool,
			Intermediates: intPool,
			KeyUsages:     []sm2.ExtKeyUsage{sm2.ExtKeyUsageAny},
		}, nil

	}
	return nil, errors.New("get options err!")

}

func getNodeCert() *gmtls.Certificate {
	logger.Info(cmd.Config.Node.NodeCertPath)

	if cert, err := gmtls.LoadX509KeyPair(cmd.Config.Node.NodeCertPath, cmd.Config.Node.NodeKeyPath); err != nil {
		return nil
	} else {
		return &cert
	}

}

func verifyCertificate(cert *gmtls.Certificate) error {
	certs := make([]*sm2.Certificate, len(cert.Certificate))
	for i, asn1Data := range cert.Certificate {
		cert, err := sm2.ParseCertificate(asn1Data)
		if err != nil {
			return errors.New("tls: failed to parse certificate from server: " + err.Error())
		}
		certs[i] = cert
	}
	opts, _ := getVerifyOptions()
	for i, cert := range certs {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}
	_, err := certs[0].Verify(*opts)
	return err

}

func CertLocalVerify() error {
	cert := getNodeCert()
	if cert == nil {
		return errors.New("get cert err!")
	}

	return verifyCertificate(cert)

}

func CertPeerVerify(i string, cert *gmtls.Certificate) error {
	c, err := json.Marshal(cert)
	if err != nil {
		logger.Error(err)
		return err
	}

	lock.Lock()
	defer lock.Unlock()
	for _, v := range ips {
		val, err := certStore.db.Get([]byte(v))
		if err != nil {
			logger.Error(err)
			return err
		}
		if string(val) == string(c) {
			logger.Error("cert reuser!")
		}
	}
	err = verifyCertificate(cert)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = certStore.db.Set([]byte(i), c)
	if err != nil {
		logger.Error(err)
		return err
	}

	ips = append(ips, i)

	return nil
}

func PeerLogOut(ip string) {

	err := certStore.db.Del([]byte(ip))
	if err != nil {
		logger.Fatal(err)
	}

	lock.Lock()
	defer lock.Unlock()
	for k, v := range ips {
		for v == ip {
			kk := k + 1
			ips = append(ips[:k], ips[kk:]...)
		}
	}
}
