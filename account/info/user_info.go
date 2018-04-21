package info

import (
  "github.com/mkmueller/aes256"
)

type UserInfo struct {
  txt []byte
}

func (info *UserInfo) ToString() string {return string(info.txt)}

func GenerateUserInfo(txt []byte) *UserInfo {
  return &UserInfo {
    txt: txt,
  }
}

func (info *UserInfo) Update(new_txt []byte) {
  info.txt = new_txt
}

func (info *UserInfo) ViewRaw(ci *aes256.Cipher) ([]byte, error) {
  raw, err := ci.Decrypt(info.txt)
  return raw, err
}

func (info *UserInfo) ViewRawString(ci *aes256.Cipher) (string, error) {
  raw, err := info.ViewRaw(ci)
  if err != nil {
    return "", err
  } else {
    return string(raw), nil
  }
}
