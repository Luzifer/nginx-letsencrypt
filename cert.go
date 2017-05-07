package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/xenolf/lego/acme"
)

func createCertificate(client *acme.Client, sld string, domains []string) (bool, error) {
	certsPath := path.Join(cfg.StorageDir, "certs")
	if err := os.MkdirAll(certsPath, 0755); err != nil {
		return false, fmt.Errorf("Unable to create certs storage directory: %s", err)
	}

	if _, err := os.Stat(path.Join(certsPath, sld+".pem")); err == nil {
		if certOK, err := verifyCertificate(path.Join(certsPath, sld+".pem"), sld, domains); err != nil {
			return false, err
		} else {
			if certOK {
				return false, nil
			}
		}
	}

	certs, failures := client.ObtainCertificate(domains, false, nil, false)
	if len(failures) > 0 {
		for d, err := range failures {
			log.WithFields(log.Fields{
				"domain": d,
			}).Error(err.Error())
		}
		return false, fmt.Errorf("Errors were recorded for certificate")
	}

	pkFile, err := os.Create(path.Join(certsPath, sld+".key"))
	if err != nil {
		return false, fmt.Errorf("Unable to create private key file: %s", err)
	}
	defer pkFile.Close()

	if _, err := pkFile.Write(certs.PrivateKey); err != nil {
		return false, fmt.Errorf("Unable to write private key: %s", err)
	}

	crtFile, err := os.Create(path.Join(certsPath, sld+".pem"))
	if err != nil {
		return false, err
	}
	defer crtFile.Close()

	if _, err := fmt.Fprintf(crtFile, "%s\n%s", certs.Certificate, certs.IssuerCertificate); err != nil {
		return false, fmt.Errorf("Unable to write certificate: %s", err)
	}

	log.WithFields(log.Fields{
		"domain": sld,
	}).Infof("Wrote new certificate")

	return true, nil
}

func verifyCertificate(filename, sld string, expectedDomains []string) (bool, error) {
	cert, err := loadCertificate(filename)
	if err != nil {
		return false, err
	}

	if cert.NotAfter.Before(time.Now().Add(cfg.BufferTime)) {
		return false, nil
	}

	if !reflect.DeepEqual(domainSort(cert.DNSNames), domainSort(expectedDomains)) {
		return false, nil
	}

	log.WithFields(log.Fields{
		"domain":      sld,
		"time_remain": cert.NotAfter.Sub(time.Now()).String(),
	}).Infof("Certificate looks good")
	return true, nil
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	pemData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	return x509.ParseCertificate(block.Bytes)
}
