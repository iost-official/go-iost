package iwallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/sdk"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

// KeyPairInfo ...
type KeyPairInfo struct {
	RawKey        string `json:"raw_key,omitempty"`
	KeyType       string `json:"key_type"`
	PubKey        string `json:"public_key"`
	EncryptMethod string `json:"encrypt_method,omitempty"`
	Salt          string `json:"salt,omitempty"`
	EncryptedKey  string `json:"encrypted_key,omitempty"`
	Mac           string `json:"mac,omitempty"`
}

// NewKeyPairInfo ...
func NewKeyPairInfo(rawKey string, keyType string) (*KeyPairInfo, error) {
	kp := &KeyPairInfo{}
	kp.RawKey = rawKey
	kp.KeyType = keyType
	key, err := account.NewKeyPair(common.Base58Decode(kp.RawKey), GetSignAlgoByName(kp.KeyType))
	if err != nil {
		return nil, err
	}
	kp.PubKey = common.Base58Encode(key.Pubkey)
	return kp, nil
}

func (k *KeyPairInfo) toKeyPair() (*account.KeyPair, error) {
	return account.NewKeyPair(common.Base58Decode(k.RawKey), GetSignAlgoByName(k.KeyType))
}

func (k *KeyPairInfo) encrypt(password []byte) error {
	if k.EncryptedKey != "" {
		return nil
	}
	salt := make([]byte, 48) // encryptKey + iv + hashSalt
	if _, err := io.ReadFull(rand.Reader, salt[0:32]); err != nil {
		return fmt.Errorf("reading from crypto/rand failed: " + err.Error())
	}
	key, err := scrypt.Key(password, salt[0:32], 32768, 8, 1, 32)
	if err != nil {
		return err
	}
	aesBlock, err := aes.NewCipher(key[0:16])
	if err != nil {
		return err
	}
	stream := cipher.NewCTR(aesBlock, salt[32:48])
	inText := common.Base58Decode(k.RawKey)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	mac := common.Sha3(append(key[16:32], outText...))

	k.EncryptMethod = "v0"
	k.Salt = common.Base58Encode(salt)
	k.EncryptedKey = common.Base58Encode(outText)
	k.Mac = common.Base58Encode(mac)
	return nil
}

func (k *KeyPairInfo) decrypt(password []byte) error {
	if k.EncryptMethod != "v0" {
		return fmt.Errorf("version mismatch")
	}
	salt := common.Base58Decode(k.Salt)

	key, err := scrypt.Key(password, salt[0:32], 32768, 8, 1, 32)
	if err != nil {
		return err
	}
	aesBlock, err := aes.NewCipher(key[0:16])
	if err != nil {
		return err
	}
	stream := cipher.NewCTR(aesBlock, salt[32:48])
	inText := common.Base58Decode(k.EncryptedKey)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	mac := common.Sha3(append(key[16:32], inText...))
	if !bytes.Equal(mac, common.Base58Decode(k.Mac)) {
		return fmt.Errorf("wrong password")
	}

	k.RawKey = common.Base58Encode(outText)
	return nil
}

// AccountInfo ...
type AccountInfo struct {
	Name     string                  `json:"name"`
	Keypairs map[string]*KeyPairInfo `json:"keypairs"`
}

// NewAccountInfo ...
func NewAccountInfo() *AccountInfo {
	return &AccountInfo{Name: "", Keypairs: make(map[string]*KeyPairInfo, 0)}
}

func (a *AccountInfo) isEncrypted() bool {
	return a.Keypairs["active"].RawKey == ""
}

func (a *AccountInfo) decrypt() error {
	cnt := 0
	for cnt <= 3 {
		cnt++
		password, err := readPasswordFromStdin(false)
		if err != nil {
			return err
		}
		retry := false
		for _, kp := range a.Keypairs {
			err := kp.decrypt(password)
			if err != nil {
				if err.Error() == "wrong password" {
					fmt.Println("decrypt error:", err)
					fmt.Println("Please retry")
					retry = true
					break
				} else {
					return err
				}
			}
		}
		if !retry {
			fmt.Println("decrypt keystore succeed")
			return nil
		}
	}
	return fmt.Errorf("load key failed")
}

func (a *AccountInfo) save(encrypt bool) error {
	dir, err := getAccountDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}
	fileName := dir + "/" + a.Name + ".json"
	if encrypt {
		fmt.Println("encrypting seckey, need password")
		password, err := readPasswordFromStdin(true)
		if err != nil {
			return err
		}
		for _, k := range a.Keypairs {
			err = k.encrypt(password)
			if err != nil {
				return err
			}
			k.RawKey = ""
		}
	} else {
		for _, k := range a.Keypairs {
			k.EncryptMethod = ""
			k.Salt = ""
			k.EncryptedKey = ""
			k.Mac = ""
		}
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, data, 0400)
	return err
}

func loadAccountFromKeyPair(fileName string) (*AccountInfo, error) {
	a := NewAccountInfo()
	for _, algo := range ValidSignAlgos {
		suf := "_" + algo
		if strings.HasSuffix(fileName, suf) {
			accName, err := getAccountNameFromKeyPath(fileName, suf)
			if err != nil {
				return nil, err
			}
			a.Name = accName
			kp, err := sdk.LoadKeyPair(fileName, algo)
			if err != nil {
				return nil, err
			}
			keyPair := &KeyPairInfo{RawKey: common.Base58Encode(kp.Seckey), PubKey: common.Base58Encode(kp.Pubkey), KeyType: algo}
			a.Keypairs["active"] = keyPair
			a.Keypairs["owner"] = keyPair
			return a, nil
		}
	}
	return nil, fmt.Errorf("invalid key file name %v", fileName)
}

func loadAccountFromKeyStore(fileName string, ensureDecrypt bool) (*AccountInfo, error) {
	a := NewAccountInfo()
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, a)
	if err != nil {
		return nil, err
	}
	if ensureDecrypt {
		if a.isEncrypted() {
			err := a.decrypt()
			if err != nil {
				return nil, err
			}
		}
	}
	return a, nil
}

func loadAccountFromFile(fileName string, ensureDecrypt bool) (*AccountInfo, error) {
	if strings.HasSuffix(fileName, ".json") {
		return loadAccountFromKeyStore(fileName, ensureDecrypt)
	}
	return loadAccountFromKeyPair(fileName)
}

func readPasswordFromStdin(repeat bool) ([]byte, error) {
	for {
		fmt.Print("Enter Password:  ")
		bytePassword, err := terminal.ReadPassword(syscall.Stdin)
		fmt.Println()
		if err != nil {
			return nil, err
		}
		if repeat {
			fmt.Print("Repeat Password:")
			repeat, err := terminal.ReadPassword(syscall.Stdin)
			fmt.Println()
			if err != nil {
				return nil, err
			}
			if !bytes.Equal(bytePassword, repeat) {
				fmt.Println("password not equal, retry")
				continue
			}
		}
		return bytePassword, nil
	}
}
