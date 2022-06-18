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
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

func main() {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Printf("Failed generating ECDSA key for [%v]: [%s]", elliptic.P256(), err)
	}
	storePrivateKey("cakey.pem", privKey)
	pubKey, _ := privKey.Public().(*ecdsa.PublicKey)

	template := x509Template()
	template.IsCA = true
	template.KeyUsage = x509.KeyUsageCertSign
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	template.UnknownExtKeyUsage = []asn1.ObjectIdentifier{[]int{1, 2, 3}, []int{2, 59, 1}}

	//set the organization for the subject
	subject := subjectTemplate()
	subject.Organization = []string{"test"}
	subject.CommonName = "test.example.com"

	template.Subject = subject

	//ski
	raw := elliptic.Marshal(privKey.Curve, privKey.PublicKey.X, privKey.PublicKey.Y)
	hash := sha256.New()
	hash.Write(raw)
	ski := hash.Sum(nil)

	template.SubjectKeyId = ski

	//跟证书
	caCert, err := genCertificateECDSA("ca.pem", &template, &template, pubKey, privKey)
	fmt.Println("create a cacert")

	//签发普通证书
	for i := 0; i < 10; i++ {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			continue
		}
		pub, _ := priv.Public().(*ecdsa.PublicKey)
		storePrivateKey("priv"+fmt.Sprintf("%d", i)+".pem", priv)
		SignCertificate("cert"+fmt.Sprintf("%d", i)+".pem", "test.example.com", nil, pub, caCert, privKey)
		fmt.Printf("create a keypair %d \n", i)
	}

}

// generate a signed X509 certficate using ECDSA
func genCertificateECDSA(fileName string, template, parent *x509.Certificate, pub *ecdsa.PublicKey,
	priv interface{}) (*x509.Certificate, error) {

	//create the x509 public cert
	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	if err != nil {
		return nil, err
	}

	//write cert out to file
	certFile, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	//pem encode the cert
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certFile.Close()
	if err != nil {
		return nil, err
	}

	x509Cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}
	return x509Cert, nil
}

func SignCertificate(fileName, name string, sans []string, pub *ecdsa.PublicKey,
	SignCert *x509.Certificate, Signer interface{}) (*x509.Certificate, error) {

	template := x509Template()
	template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}

	//set the organization for the subject
	subject := subjectTemplate()
	subject.CommonName = name

	template.Subject = subject
	template.DNSNames = sans

	cert, err := genCertificateECDSA(fileName, &template, SignCert, pub, Signer)

	if err != nil {
		return nil, err
	}
	return cert, nil
}

// default template for X509 subject
func subjectTemplate() pkix.Name {
	return pkix.Name{
		// Country:  []string{"US"},
		// Locality: []string{"San Francisco"},
		// Province: []string{"California"},

		Country: []string{"China"},
		ExtraNames: []pkix.AttributeTypeAndValue{
			{
				Type:  []int{2, 5, 4, 42},
				Value: "Gopher",
			},
			// This should override the Country, above.
			{
				Type:  []int{2, 5, 4, 6},
				Value: "NL",
			},
		},
	}
}

// default template for X509 certificates
func x509Template() x509.Certificate {
	//generate a serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	now := time.Now()
	//basic template to use
	x509 := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             now,
		NotAfter:              now.Add(3650 * 24 * time.Hour), //~ten years
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
	}
	return x509
}

type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

type pkcs8Info struct {
	Version             int
	PrivateKeyAlgorithm []asn1.ObjectIdentifier
	PrivateKey          []byte
}

func storePrivateKey(alias string, privateKey *ecdsa.PrivateKey) error {
	rawKey, err := privateKeyToPEM(privateKey)
	if err != nil {
		return fmt.Errorf("Failed converting private key to PEM [%s]: [%s]", alias, err)
	}
	err = ioutil.WriteFile(alias, rawKey, 0700)
	if err != nil {
		return fmt.Errorf("Failed storing private key [%s]: [%s]", alias, err)
	}
	return nil
}

func privateKeyToPEM(privateKey interface{}) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("Invalid key. It must be different from nil.")
	}
	switch k := privateKey.(type) {
	case *ecdsa.PrivateKey:
		// oidNamedCurveP256
		var oidNamedCurve = asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}
		var oidPublicKeyECDSA = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}

		// based on https://golang.org/src/crypto/x509/sec1.go
		privateKeyBytes := k.D.Bytes()
		paddedPrivateKey := make([]byte, (k.Curve.Params().N.BitLen()+7)/8)
		copy(paddedPrivateKey[len(paddedPrivateKey)-len(privateKeyBytes):], privateKeyBytes)
		// omit NamedCurveOID for compatibility as it's optional
		asn1Bytes, err := asn1.Marshal(ecPrivateKey{
			Version:    1,
			PrivateKey: paddedPrivateKey,
			PublicKey:  asn1.BitString{Bytes: elliptic.Marshal(k.Curve, k.X, k.Y)},
		})

		if err != nil {
			return nil, fmt.Errorf("error marshaling EC key to asn1 [%s]", err)
		}

		var pkcs8Key pkcs8Info
		pkcs8Key.Version = 0
		pkcs8Key.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 2)
		pkcs8Key.PrivateKeyAlgorithm[0] = oidPublicKeyECDSA
		pkcs8Key.PrivateKeyAlgorithm[1] = oidNamedCurve
		pkcs8Key.PrivateKey = asn1Bytes

		pkcs8Bytes, err := asn1.Marshal(pkcs8Key)
		if err != nil {
			return nil, fmt.Errorf("error marshaling EC key to asn1 [%s]", err)
		}
		return pem.EncodeToMemory(
			&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: pkcs8Bytes,
			},
		), nil

	case *rsa.PrivateKey:
		if k == nil {
			return nil, errors.New("Invalid rsa private key. It must be different from nil.")
		}
		raw := x509.MarshalPKCS1PrivateKey(k)

		return pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: raw,
			},
		), nil
	default:
		return nil, errors.New("Invalid key type. It must be *ecdsa.PrivateKey or *rsa.PrivateKey")
	}
}
