package keyprovider

import (
	"errors"
	"os"
	"time"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"log/slog"
)

type KeystoreProvider struct {
	ks       keystore.KeyStore
	filePath string
}

func NewKeystoreProvider(filePath, password string) (*KeystoreProvider, error) {
	ks := keystore.New()

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &KeystoreProvider{ks: ks, filePath: filePath}, nil
		}
		return nil, err
	}
	defer f.Close()

	err = ks.Load(f, []byte(password))
	if err != nil {
		return nil, err
	}

	return &KeystoreProvider{ks: ks, filePath: filePath}, nil
}

func (k *KeystoreProvider) GetPrivateKey(keyTag uint8) ([]byte, error) {
	alias, err := getAlias(keyTag)
	if err != nil {
		return nil, err
	}

	entry, err := k.ks.GetPrivateKeyEntry(alias, []byte{})
	if err != nil {
		return nil, err
	}
	return entry.PrivateKey, nil
}

func (k *KeystoreProvider) HasKey(keyTag uint8) (bool, error) {
	alias, err := getAlias(keyTag)
	if err != nil {
		return false, err
	}
	return k.ks.IsPrivateKeyEntry(alias), nil
}

func (k *KeystoreProvider) AddKey(keyTag uint8, privateKey []byte, password string) error {
	exists, err := k.HasKey(keyTag)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("key already exists")
	}

	alias, err := getAlias(keyTag)
	if err != nil {
		return err
	}

	err = k.ks.SetPrivateKeyEntry(alias, keystore.PrivateKeyEntry{
		CreationTime:     time.Now(),
		PrivateKey:       privateKey,
		CertificateChain: nil,
	}, []byte{})
	if err != nil {
		return err
	}

	err = k.dump(password)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeystoreProvider) DeleteKey(keyTag uint8, password string) error {
	exists, err := k.HasKey(keyTag)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("key does not exist")
	}

	alias, err := getAlias(keyTag)
	if err != nil {
		return err
	}

	k.ks.DeleteEntry(alias)

	err = k.dump(password)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeystoreProvider) dump(password string) error {
	f, err := os.Create(k.filePath)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			slog.Error(err.Error())
		}
	}()

	err = k.ks.Store(f, []byte(password))
	if err != nil {
		return err
	}

	return nil
}
