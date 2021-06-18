package iwallet

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iost-official/go-iost/v3/account"
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/sdk"
	. "github.com/iost-official/go-iost/v3/sdk"
	"github.com/spf13/cobra"
)

var (
	ownerKey         string
	activeKey        string
	initialRAM       int64
	initialBalance   int64
	initialGasPledge int64
)

type acc struct {
	Name    string
	KeyPair *key
}

type accounts struct {
	Dir     string
	Account []*acc
}

// accountCmd represents the account command.
var accountCmd = &cobra.Command{
	Use:     "account",
	Aliases: []string{"acc"},
	Short:   "KeyPair manager",
	Long:    `Manage account in local storage`,
}
var updateCmd = &cobra.Command{
	Use:   "update keypair",
	Short: "Update account keypair",
	Long:  `Update account keypair`,
	Example: `  iwallet account update --account test0
  iwallet account update --account test0 --owner 7Z9US64vfcyopQpyEwV1FF52HTB8maEacjU4SYeAUrt1 --active 7Z9US64vfcyopQpyEwV1FF52HTB8maEacjU4SYeAUrt1`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set account since making actions needs accountName.
		// TODO: make these lines more clean...
		signPerm = "owner"
		err := initAccountForSDK(iwalletSDK)
		if err != nil {
			return err
		}
		signers = append(signers, accountName+"@owner")

		accInfo, err := getAccountInfoFromArgs(accountName)
		if err != nil {
			return err
		}
		akey := accInfo.Keypairs["active"].PubKey
		okey := accInfo.Keypairs["owner"].PubKey
		actions, err := iwalletSDK.UpdateAccountKeysActions(accountName, okey, akey)
		if err != nil {
			return err
		}
		_, err = processActions(actions)
		if err != nil {
			return err
		}
		return postAccountUpdateHandler(accountName, accInfo)
	},
}

func getAccountInfoFromArgs(name string) (*AccountInfo, error) {
	accInfo := NewAccountInfo()
	accInfo.Name = name
	if ownerKey == "" && activeKey == "" {
		newKp, err := account.NewKeyPair(nil, sdk.GetSignAlgoByName(signAlgo))
		if err != nil {
			err = fmt.Errorf("failed to create key pair: %v", err)
			return nil, err
		}
		kp := &KeyPairInfo{
			RawKey:  common.Base58Encode(newKp.Seckey),
			PubKey:  common.Base58Encode(newKp.Pubkey),
			KeyType: signAlgo,
		}
		accInfo.Keypairs["active"] = kp
		accInfo.Keypairs["owner"] = kp
		return accInfo, nil
	} else if sdk.CheckPubKey(ownerKey) && sdk.CheckPubKey(activeKey) {
		accInfo.Keypairs["active"] = &KeyPairInfo{
			RawKey:  "",
			PubKey:  activeKey,
			KeyType: signAlgo,
		}
		accInfo.Keypairs["owner"] = &KeyPairInfo{
			RawKey:  "",
			PubKey:  ownerKey,
			KeyType: signAlgo,
		}
		return accInfo, nil
	} else {
		return nil, fmt.Errorf("key provided but not valid")
	}
}

func postAccountUpdateHandler(newName string, accInfo *AccountInfo) error {
	realUpdated := outputKeyFile == "" && !tryTx
	if !realUpdated {
		return nil
	}
	// step1 print new account info fetched from chain
	if realUpdated && checkResult {
		info, err := iwalletSDK.GetAccountInfo(newName)
		if err != nil {
			return fmt.Errorf("failed to get account info: %v", err)
		}
		fmt.Println("Account info of <", newName, ">:")
		fmt.Println(sdk.MarshalTextString(info))
	}
	// step2 print new account info locally
	akey := accInfo.Keypairs["active"].PubKey
	okey := accInfo.Keypairs["owner"].PubKey
	fmt.Println("The IOST account ID is:", newName)
	fmt.Println("Owner permission key:", okey)
	fmt.Println("Active permission key:", akey)
	if realUpdated {
		// step3 save account info
		if accInfo.Keypairs["active"].RawKey != "" || accInfo.Keypairs["owner"].RawKey != "" {
			err := saveAccount(accInfo, encrypt)
			if err != nil {
				return fmt.Errorf("failed to save account: %v %v", err, accInfo)
			}
		}
	}
	return nil
}

var createCmd = &cobra.Command{
	Use:   "create accountName",
	Short: "Create an account on blockchain",
	Long:  `Create an account on blockchain`,
	Example: `  iwallet account create test1 --account test0
  iwallet account create test2 --account test0 --initial_balance 0 --initial_gas_pledge 10 --initial_ram 0
  iwallet account create test3 --account test0 --owner 7Z9US64vfcyopQpyEwV1FF52HTB8maEacjU4SYeAUrt1 --active 7Z9US64vfcyopQpyEwV1FF52HTB8maEacjU4SYeAUrt1`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "accountName"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err error
		)

		newName := args[0]
		if strings.ContainsAny(newName, `?*:|/\"`) || len(newName) < 5 || len(newName) > 11 {
			return fmt.Errorf("invalid account name")
		}
		accInfo, err := getAccountInfoFromArgs(newName)
		if err != nil {
			return err
		}
		akey := accInfo.Keypairs["active"].PubKey
		okey := accInfo.Keypairs["owner"].PubKey
		// Set account since making actions needs accountName.
		err = initAccountForSDK(iwalletSDK)
		if err != nil {
			return err
		}
		actions, err := iwalletSDK.CreateNewAccountActions(newName, okey, akey, initialGasPledge, initialRAM, initialBalance)
		if err != nil {
			return err
		}
		_, err = processActions(actions)
		if err != nil {
			return err
		}
		return postAccountUpdateHandler(newName, accInfo)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view [accountName]",
	Short: "View account by name or omit to show all accounts",
	Long:  `View account by name or omit to show all accounts`,
	Example: `  iwallet account view test0
  iwallet account view`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a := accounts{}
		a.Dir = defaultFileAccountStore.AccountDir
		addAcc := func(ac *AccountInfo) {
			var k key
			k.Algorithm = ac.Keypairs["active"].KeyType
			k.Pubkey = ac.Keypairs["active"].PubKey
			if ac.IsEncrypted() {
				k.Seckey = "---encrypted secret key---"
			} else {
				k.Seckey = ac.Keypairs["active"].RawKey
			}
			a.Account = append(a.Account, &acc{ac.Name, &k})
		}
		if len(args) < 1 {
			accs, err := defaultFileAccountStore.ListAccounts()
			if err != nil {
				return err
			}
			for _, ac := range accs {
				addAcc(ac)
			}
		} else {
			name := args[0]
			ac, err := defaultFileAccountStore.LoadAccount(name)
			if err != nil {
				return err
			}
			addAcc(ac)
		}
		info, err := json.MarshalIndent(a, "", "    ")
		if err != nil {
			return err
		}
		fmt.Println(string(info))
		return nil
	},
}

var encrypt bool
var importCmd = &cobra.Command{
	Use:   "import accountName accountPrivateKey",
	Short: "Import an account by name and private key",
	Long:  `Import an account by name and private key`,
	Example: `  iwallet account import test0 XXXXXXXXXXXXXXXXXXXXX
  iwallet account import test0 active:XXXXXXXXXXXXXXXXXXXXX,owner:YYYYYYYYYYYYYYYYYYYYYYYY`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "accountName", "accountPrivateKey"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		acc := AccountInfo{Name: name, Keypairs: make(map[string]*KeyPairInfo)}
		keys := strings.Split(args[1], ",")
		if len(keys) == 1 {
			key := keys[0]
			if len(strings.Split(key, ":")) != 1 {
				return fmt.Errorf("importing one key need not specifying permission")
			}
			kp, err := NewKeyPairInfo(key, signAlgo)
			if err != nil {
				return err
			}
			acc.Keypairs["active"] = kp
			acc.Keypairs["owner"] = kp
		} else {
			for _, permAndKey := range keys {
				splits := strings.Split(permAndKey, ":")
				if len(splits) != 2 {
					return fmt.Errorf("importing more than one keys need specifying permissions")
				}
				kp, err := NewKeyPairInfo(splits[1], signAlgo)
				if err != nil {
					return err
				}
				acc.Keypairs[splits[0]] = kp
			}
		}
		err := saveAccount(&acc, encrypt)
		if err != nil {
			return fmt.Errorf("failed to save account: %v", err)
		}
		fmt.Printf("import account %v done\n", name)
		return nil
	},
}

var dumpKeyCmd = &cobra.Command{
	Use:     "dumpkey accountName",
	Short:   "Print private key of the account to stdout",
	Long:    "Print private key of the account to stdout",
	Example: `  iwallet account dumpkey test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "accountName"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		accountName = args[0]
		acc, err := loadAccount(true)
		if err != nil {
			return err
		}
		for k, v := range acc.Keypairs {
			fmt.Printf("%v:%v\n", k, v.RawKey)
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete accountName",
	Aliases: []string{"del"},
	Short:   "Delete an account by name",
	Long:    `Delete an account by name`,
	Example: `  iwallet account delete test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "accountName"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		err := defaultFileAccountStore.DeleteAccount(name)
		if err != nil {
			fmt.Println("Account <", name, "> does not exist:", err)
		} else {
			fmt.Println("Successfully deleted <", name, ">.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
	accountCmd.PersistentFlags().BoolVarP(&encrypt, "encrypt", "", false, "whether to encrypt local key file")

	accountCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&ownerKey, "owner", "", "", "owner key")
	createCmd.Flags().StringVarP(&activeKey, "active", "", "", "active key")
	createCmd.Flags().Int64VarP(&initialRAM, "initial_ram", "", 1024, "buy $initial_ram bytes ram for the new account")
	createCmd.Flags().Int64VarP(&initialGasPledge, "initial_gas_pledge", "", 10, "pledge $initial_gas_pledge IOSTs for the new account")
	createCmd.Flags().Int64VarP(&initialBalance, "initial_balance", "", 0, "transfer $initial_balance IOSTs to the new account")

	accountCmd.AddCommand(importCmd)
	accountCmd.AddCommand(viewCmd)
	accountCmd.AddCommand(deleteCmd)
	accountCmd.AddCommand(dumpKeyCmd)
	accountCmd.AddCommand(updateCmd)
}
