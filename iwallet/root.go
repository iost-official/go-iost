package iwallet

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iost-official/go-iost/v3/sdk"
)

var cfgFile string
var startTime time.Time

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:           "iwallet",
	Short:         "IOST client",
	Long:          `An IOST RPC client`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		startTime = time.Now()
		iwalletSDK = sdk.NewIOSTDevSDK()
		iwalletSDK.SetChainID(chainID)
		if !strings.Contains(server, ":") {
			// use default port
			server += ":30002"
		}
		iwalletSDK.SetServer(server)
		iwalletSDK.SetVerbose(verbose)
		iwalletSDK.SetCheckResult(checkResult, checkResultDelay, checkResultMaxRetry)
		limit, err := ParseAmountLimit(amountLimit)
		if err != nil {
			return fmt.Errorf("invalid amount limit %v: %v", amountLimit, err)
		}
		if !(0 < expiration && expiration <= 90) {
			return fmt.Errorf("expiration should be in (0, 90]")
		}
		iwalletSDK.SetTxInfo(gasLimit, gasRatio, expiration, delaySecond, limit)
		iwalletSDK.SetUseLongestChain(useLongestChain)
		initFileAccountStore()
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if elapsedTime && (len(cmd.Use) < 4 || cmd.Use[:4] != "help") {
			fmt.Println("Executed in", time.Since(startTime))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("\033[31mERROR:\033[m", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "configuration file (default $HOME/.iwallet.yaml)")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", true, "print verbose information")
	rootCmd.PersistentFlags().BoolVarP(&elapsedTime, "elapsed_time", "", false, "print elapsed time")
	rootCmd.PersistentFlags().StringVarP(&accountName, "account", "a", "", "which account to use")
	rootCmd.PersistentFlags().StringVarP(&accountDir, "account_dir", "", "", "$(account_dir)/.iwallet will be used to save accounts (default $HOME/.iwallet)")
	rootCmd.PersistentFlags().StringVarP(&keyFile, "key_file", "k", "", "the json file that stores your account name and keypair")
	rootCmd.PersistentFlags().StringVarP(&outputKeyFile, "output_key_file", "", "", "the json file used to save your newly created account name and keypair, only used in 'iwallet account create/import' command")
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "localhost:30002", "set server of this client")
	rootCmd.PersistentFlags().BoolVarP(&useLongestChain, "use_longest", "", false, "get info on longest chain")
	rootCmd.PersistentFlags().BoolVarP(&checkResult, "check_result", "", true, "check publish/call status after sending to chain")
	rootCmd.PersistentFlags().BoolVarP(&tryTx, "try_tx", "", false, "Executes a new call immediately without creating a transaction on the block chain")
	rootCmd.PersistentFlags().Float32VarP(&checkResultDelay, "check_result_delay", "", 3, "rpc checking will occur at [checkResultDelay] seconds after sending to chain.")
	rootCmd.PersistentFlags().Int32VarP(&checkResultMaxRetry, "check_result_max_retry", "", 30, "max times to call grpc to check tx status")
	rootCmd.PersistentFlags().StringVarP(&signAlgo, "sign_algo", "", "ed25519", "sign algorithm")
	rootCmd.PersistentFlags().StringSliceVarP(&signers, "signers", "", []string{}, "additional signers")
	rootCmd.PersistentFlags().Float64VarP(&gasLimit, "gas_limit", "l", 1000000, "gas limit for a transaction")
	rootCmd.PersistentFlags().Float64VarP(&gasRatio, "gas_ratio", "p", 1.0, "gas ratio for a transaction")
	rootCmd.PersistentFlags().StringVarP(&amountLimit, "amount_limit", "", "*:unlimited", "amount limit for one transaction, eg iost:300.00|ram:2000")
	rootCmd.PersistentFlags().Int64VarP(&expiration, "expiration", "e", 90, "expiration time for a transaction in seconds")
	rootCmd.PersistentFlags().Uint32VarP(&chainID, "chain_id", "", uint32(1024), "chain id which distinguishes different network")
	rootCmd.PersistentFlags().StringVarP(&txTime, "tx_time", "", "", fmt.Sprintf("use the special tx time instead of now, format: %v", time.Now().Format(time.RFC3339)))
	rootCmd.PersistentFlags().Uint32VarP(&txTimeDelay, "tx_time_delay", "", 0, "delay the tx time from now")
	rootCmd.PersistentFlags().StringVarP(&signPerm, "sign_permission", "", "active", "permission used to sign transactions")
	rootCmd.PersistentFlags().StringVarP(&outputTxFile, "output", "o", "", "output json file to save transaction request")
	// the following three flags are used for multiple signature
	rootCmd.PersistentFlags().StringSliceVarP(&signKeyFiles, "sign_key_files", "", []string{}, "optional private key files used for signing, split by comma")
	rootCmd.PersistentFlags().StringSliceVarP(&signatureFiles, "signature_files", "", []string{}, "optional signature files, split by comma")
	rootCmd.PersistentFlags().BoolVarP(&asPublisherSign, "as_publisher_sign", "", false, "use signKeyFiles/signatureFiles for publisher sign")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".iwallet" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".iwallet")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

var iwalletSDK *sdk.IOSTDevSDK

var (
	server        string
	accountName   string
	accountDir    string
	keyFile       string
	outputKeyFile string
	signAlgo      string
	signers       []string
	signPerm      string

	gasLimit     float64
	gasRatio     float64
	expiration   int64
	amountLimit  string
	delaySecond  int64
	txTime       string
	txTimeDelay  uint32
	outputTxFile string

	// Used for multi sig.
	signKeyFiles    []string
	signatureFiles  []string
	asPublisherSign bool

	tryTx               bool
	checkResult         bool
	checkResultDelay    float32
	checkResultMaxRetry int32
	useLongestChain     bool

	verbose     bool
	elapsedTime bool

	chainID uint32
)
