package account

import (
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
)

var (
//GenesisAccount = map[string]int64{
//	"IOST7j5ynm6UXmJqY3GJp4RchiEWxjoPvifqgQVVAUaKV9sPLpx719": 3400000000,
//	"IOST5QWmFgVpXwz7vgQmPd4qUHDbexKeK523PtcMDh8geyM131niP2": 3200000000,
//	"IOST7koa9yKFh4j45t97xwMBzVKrJw1q8mnixRmsdJx4JftnLo52tH": 3100000000,
//	"IOST5dFQcKgmU3ZaKDgfVuX2vVCgm3W9W1GyJWyhEtgfibK6nM6Hsn": 3000000000,
//	"IOST7mo771oBeAuYNp7b8VLhkP7FHxcub4HeWx3VzqsLvmNKxatQ8c": 2900000000,
//	"IOST5XxHPKQ5gdvEocak4rjeDKgLh3uYumrT1iUFoR82vrHkMWNz4p": 2800000000,
//	"IOST8SJYAr9ig81JPvjnyeDsPX9fbDuNGB4ECWKTedvMBN4fzZiG7F": 2600000000,
//}
//WitnessList = []string{
//	"IOST7j5ynm6UXmJqY3GJp4RchiEWxjoPvifqgQVVAUaKV9sPLpx719",
//	"IOST5QWmFgVpXwz7vgQmPd4qUHDbexKeK523PtcMDh8geyM131niP2",
//	"IOST7koa9yKFh4j45t97xwMBzVKrJw1q8mnixRmsdJx4JftnLo52tH",
//	"IOST5dFQcKgmU3ZaKDgfVuX2vVCgm3W9W1GyJWyhEtgfibK6nM6Hsn",
//	"IOST7mo771oBeAuYNp7b8VLhkP7FHxcub4HeWx3VzqsLvmNKxatQ8c",
//	"IOST5XxHPKQ5gdvEocak4rjeDKgLh3uYumrT1iUFoR82vrHkMWNz4p",
//	"IOST8SJYAr9ig81JPvjnyeDsPX9fbDuNGB4ECWKTedvMBN4fzZiG7F",
//}

//local net
//GenesisAccount = map[string]int64{
//	"IOST5FhLBhVXMnwWRwhvz5j9NyWpBSchAMzpSMZT21xZqT8w7icwJ5": 13400000000, // seckey:BCV7fV37aSWNx1N1Yjk3TdQXeHMmLhyqsqGms1PkqwPT
//	"IOST6Jymdka3EFLAv8954MJ1nBHytNMwBkZfcXevE2PixZHsSrRkbR": 13200000000, // seckey:2Hoo4NAoFsx9oat6qWawHtzqFYcA3VS7BLxPowvKHFPM
//	"IOST7gKuvHVXtRYupUixCcuhW95izkHymaSsgKTXGDjsyy5oTMvAAm": 13100000000, // seckey:6nMnoZqgR7Nvs6vBHiFscEtHpSYyvwupeDAyfke12J1N
//}
//
//WitnessList = []string{
//	"IOST5FhLBhVXMnwWRwhvz5j9NyWpBSchAMzpSMZT21xZqT8w7icwJ5",
//	"IOST6Jymdka3EFLAv8954MJ1nBHytNMwBkZfcXevE2PixZHsSrRkbR",
//	"IOST7gKuvHVXtRYupUixCcuhW95izkHymaSsgKTXGDjsyy5oTMvAAm",
//}
)

type Account struct {
	ID        string
	Algorithm crypto.Algorithm
	Pubkey    []byte
	Seckey    []byte
}

func NewAccount(seckey []byte, algo crypto.Algorithm) (*Account, error) {
	if seckey == nil {
		seckey = algo.GenSeckey()
	}
	if (len(seckey) != 32 && algo == crypto.Secp256k1) ||
		(len(seckey) != 64 && algo == crypto.Ed25519) {
		return nil, fmt.Errorf("seckey length error")
	}
	pubkey := algo.GetPubkey(seckey)
	id := GetIDByPubkey(pubkey)

	account := &Account{
		ID:        id,
		Algorithm: algo,
		Pubkey:    pubkey,
		Seckey:    seckey,
	}
	return account, nil
}

func (a *Account) Sign(info []byte) *crypto.Signature {
	return crypto.NewSignature(a.Algorithm, info, a.Seckey)
}

func GetIDByPubkey(pubkey []byte) string {
	return "IOST" + common.Base58Encode(append(pubkey, common.Parity(pubkey)...))
}

func GetPubkeyByID(ID string) []byte {
	b := common.Base58Decode(ID[4:])
	return b[:len(b)-4]
}
