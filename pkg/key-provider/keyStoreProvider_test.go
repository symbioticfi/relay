package keyprovider

import (
	"bytes"
	"os"
	"testing"
)

func TestNewKeystore(t *testing.T) {
	path := "./TMP-keystore"
	password := "password"
	defer func() {
		_ = os.Remove(path)
	}()

	_, err := NewKeystoreProvider(path, password)
	if err != nil {
		t.Errorf("error creating signer: %v", err)
	}
}

func TestAddKey(t *testing.T) {
	path := "./TMP-keystore"
	password := "password"
	defer func() {
		_ = os.Remove(path)
	}()

	kp, err := NewKeystoreProvider(path, password)
	if err != nil {
		t.Errorf("error creating signer: %v", err)
	}

	pk := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	err = kp.AddKey(15, pk, password)
	if err != nil {
		t.Errorf("error adding key: %v", err)
	}
}

func TestCreateAndReopen(t *testing.T) {
	path := "./TMP-keystore"
	password := "password"

	defer func() {
		_ = os.Remove(path)
	}()

	kp, err := NewKeystoreProvider(path, password)
	if err != nil {
		t.Errorf("error creating signer: %v", err)
	}

	pk := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	err = kp.AddKey(15, pk, password)
	if err != nil {
		t.Errorf("error adding key: %v", err)
	}

	kp, err = NewKeystoreProvider(path, password)
	if err != nil {
		t.Errorf("error laoding keystore: %v", err)
	}
	exists, err := kp.HasKey(15)
	if err != nil {
		t.Errorf("error checking key: %v", err)
	}

	if !exists {
		t.Errorf("key not found in keystore")
	}

	storedPk, err := kp.GetPrivateKey(15)
	if err != nil {
		t.Errorf("error getting key: %v", err)
	}

	if !bytes.Equal(storedPk, pk) {
		t.Errorf("stored private key not equal to original private key")
	}
}
