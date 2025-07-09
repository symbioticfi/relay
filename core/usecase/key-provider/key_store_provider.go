package keyprovider

import (
	"errors"
	"log/slog"
	"middleware-offchain/core/usecase/crypto"
	"os"
	"path/filepath"
	"time"

	"github.com/pavlo-v-chernykh/keystore-go/v4"

	"middleware-offchain/core/entity"
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

func (k *KeystoreProvider) GetAliases() []string {
	return k.ks.Aliases()
}

func (k *KeystoreProvider) GetPrivateKey(keyTag entity.KeyTag) (crypto.PrivateKey, error) {
	alias, err := getAlias(keyTag)
	if err != nil {
		return nil, err
	}

	entry, err := k.ks.GetPrivateKeyEntry(alias, []byte{})
	if err != nil {
		return nil, err
	}
	return crypto.NewPrivateKey(keyTag, entry.PrivateKey)
}

func (k *KeystoreProvider) HasKey(keyTag entity.KeyTag) (bool, error) {
	alias, err := getAlias(keyTag)
	if err != nil {
		return false, err
	}
	return k.ks.IsPrivateKeyEntry(alias), nil
}

func (k *KeystoreProvider) AddKey(keyTag entity.KeyTag, privateKey crypto.PrivateKey, password string, force bool) error {
	exists, err := k.HasKey(keyTag)
	if err != nil {
		return err
	}

	if exists && !force {
		return errors.New("key already exists")
	}

	alias, err := getAlias(keyTag)
	if err != nil {
		return err
	}

	err = k.ks.SetPrivateKeyEntry(alias, keystore.PrivateKeyEntry{
		CreationTime:     time.Now(),
		PrivateKey:       privateKey.Bytes(),
		CertificateChain: nil,
	}, []byte{})
	if err != nil {
		return err
	}

	err = k.dump(password)
	if err != nil {
		return err
	}

	if exists {
		slog.Info("Key was updated!")
	}

	return nil
}

func (k *KeystoreProvider) DeleteKey(keyTag entity.KeyTag, password string) error {
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
	dir := filepath.Dir(k.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(k.filePath)
	if err != nil {
		slog.Error("Failed to create file", "err", err.Error(), "path", k.filePath)
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("Failed to close file", "err", err.Error())
		}
	}()

	err = k.ks.Store(f, []byte(password))
	if err != nil {
		return err
	}

	return nil
}
