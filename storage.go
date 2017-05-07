package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"os"
	"path"

	"github.com/xenolf/lego/acme"
)

const (
	userStorageFile = "registration.json"
	rsaKeySize      = 2048
)

type user struct {
	Email        string
	Registration *acme.RegistrationResource
	key          *rsa.PrivateKey
	Key          []byte
}

func (u user) GetEmail() string {
	return u.Email
}
func (u user) GetRegistration() *acme.RegistrationResource {
	return u.Registration
}
func (u user) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func loadOrCreateUser() (*user, error) {
	storageFilePath := path.Join(cfg.StorageDir, userStorageFile)

	if _, err := os.Stat(storageFilePath); err != nil {
		privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
		if err != nil {
			return nil, err
		}

		u := &user{
			Email: cfg.Email,
			key:   privateKey,
		}

		return u, nil
	}

	storageFile, err := os.Open(storageFilePath)
	if err != nil {
		return nil, err
	}
	defer storageFile.Close()
	defer os.Chmod(storageFilePath, 0600)

	u := &user{}
	if err := json.NewDecoder(storageFile).Decode(u); err != nil {
		return nil, err
	}

	key, err := x509.ParsePKCS1PrivateKey(u.Key)
	u.key = key

	return u, nil
}

func (u *user) Save() error {
	storageFilePath := path.Join(cfg.StorageDir, userStorageFile)

	if err := os.MkdirAll(cfg.StorageDir, 0755); err != nil {
		return err
	}

	storageFile, err := os.Create(storageFilePath)
	if err != nil {
		return err
	}
	defer storageFile.Close()

	u.Key = x509.MarshalPKCS1PrivateKey(u.key)
	return json.NewEncoder(storageFile).Encode(u)
}
