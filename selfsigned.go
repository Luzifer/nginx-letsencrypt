package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"time"
)

func generateSelfsignedCert(filename string) error {
	template := &x509.Certificate{
		IsCA: true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"Earth"},
			Organization: []string{"Mother Nature"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 5, 5),
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// generate private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		fmt.Println(err)
	}

	publickey := &privatekey.PublicKey

	// create a self-signed certificate. template = parent
	var parent = template
	cert, err := x509.CreateCertificate(rand.Reader, template, parent, publickey, privatekey)

	if err != nil {
		fmt.Println(err)
	}

	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		return err
	}

	// save private key
	pkey := x509.MarshalPKCS1PrivateKey(privatekey)
	if err := ioutil.WriteFile(filename+".key", pem.EncodeToMemory(&pem.Block{Bytes: pkey, Type: "RSA PRIVATE KEY"}), 0644); err != nil {
		return err
	}

	// save public key
	//pubkey, _ := x509.MarshalPKIXPublicKey(publickey)
	//ioutil.WriteFile("public.key", pubkey, 0644)
	//fmt.Println("public key saved to public.key")

	// save cert
	if err := ioutil.WriteFile(filename+".pem", pem.EncodeToMemory(&pem.Block{Bytes: cert, Type: "TRUSTED CERTIFICATE"}), 0644); err != nil {
		return err
	}

	return nil
}

func ensureCertFiles(nameGroups map[string][]string) error {
	certsPath := path.Join(cfg.StorageDir, "certs")
	for sld := range nameGroups {
		certPath := path.Join(certsPath, sld)
		if _, err := os.Stat(certPath + ".pem"); err != nil {
			if err := generateSelfsignedCert(certPath); err != nil {
				return err
			}
		}
	}

	return nil
}
