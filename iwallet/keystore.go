package iwallet

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/sdk"
	. "github.com/iost-official/go-iost/v3/sdk"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/term"
)

var defaultFileAccountStore *sdk.FileAccountStore

func initFileAccountStore() {
	if accountDir == "" {
		home, err := homedir.Dir()
		if err != nil {
			panic(fmt.Errorf("cannot get home dir %v", err))
		}
		accountDir = path.Join(home, ".iwallet")
	}
	defaultFileAccountStore = NewFileAccountStore(accountDir)
}

func encryptAccount(a *AccountInfo) error {
	fmt.Println("encrypting seckey, need password")
	password, err := readPasswordFromStdin(true)
	if err != nil {
		return err
	}
	return a.Encrypt(password)
}

// nolint
func saveAccountTo(a *AccountInfo, fileName string, encrypt bool) error {
	if encrypt {
		err := encryptAccount(a)
		if err != nil {
			return err
		}
	}
	return a.SaveTo(fileName)
}

func saveAccount(a *AccountInfo, encrypt bool) error {
	if encrypt {
		err := encryptAccount(a)
		if err != nil {
			return err
		}
	}
	if outputKeyFile != "" {
		if accountDir != "" {
			ilog.Warn("--key_file is set, so --account_dir will be ignored")
		}
		return a.SaveTo(outputKeyFile)
	}
	return defaultFileAccountStore.SaveAccount(a)
}

func ensureDecryptAccount(a *AccountInfo) error {
	if !a.IsEncrypted() {
		return nil
	}
	password, err := readPasswordFromStdin(false)
	if err != nil {
		return err
	}
	return a.Decrypt(password)
}

func loadAccountFrom(fileName string, ensureDecrypt bool) (*AccountInfo, error) {
	a, err := LoadAccountFrom(fileName)
	if err != nil {
		return nil, err
	}
	if ensureDecrypt {
		err = ensureDecryptAccount(a)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

func loadAccount(ensureDecrypt bool) (*AccountInfo, error) {
	var acc *AccountInfo
	var err error
	if keyFile != "" {
		if accountDir != "" {
			ilog.Warn("--key_file is set, so --account_dir will be ignored")
		}
		acc, err = LoadAccountFrom(keyFile)
		if err != nil {
			return nil, err
		}
		if accountName != "" && acc.Name != accountName {
			return nil, fmt.Errorf("inconsistent account: %s from cmd args VS %s from key file", accountName, acc.Name)
		}
		if accountName == "" {
			accountName = acc.Name
		}
	} else {
		acc, err = defaultFileAccountStore.LoadAccount(accountName)
	}
	if ensureDecrypt {
		err = ensureDecryptAccount(acc)
		if err != nil {
			return nil, err
		}
	}
	return acc, err
}

func initAccountForSDK(s *sdk.IOSTDevSDK) error {
	a, err := loadAccount(true)
	if err != nil {
		return err
	}
	keyPair, err := a.GetKeyPair(signPerm)
	if err != nil {
		return err
	}
	s.SetAccount(a.Name, keyPair)
	s.UseAccount(a.Name)
	return nil
}

func readPassword(prompt string) (pw []byte, err error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		fmt.Fprint(os.Stderr, prompt)
		pw, err = term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		return
	}

	var b [1]byte
	for {
		n, err := os.Stdin.Read(b[:])
		// terminal.ReadPassword discards any '\r', so we do the same
		if n > 0 && b[0] != '\r' {
			if b[0] == '\n' {
				return pw, nil
			}
			pw = append(pw, b[0])
			// limit size, so that a wrong input won't fill up the memory
			if len(pw) > 1024 {
				err = errors.New("password too long")
			}
		}
		if err != nil {
			// terminal.ReadPassword accepts EOF-terminated passwords
			// if non-empty, so we do the same
			if err == io.EOF && len(pw) > 0 {
				err = nil
			}
			return pw, err
		}
	}
}

func readPasswordFromStdin(repeat bool) ([]byte, error) {
	for {
		bytePassword, err := readPassword("Enter Password:  ")
		if err != nil {
			return nil, err
		}
		if repeat {
			repeat, err := readPassword("Enter Password:  ")
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
