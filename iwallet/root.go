package iwallet

import (
	"fmt"
	"os"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/sdk"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:           "iwallet",
	Short:         "IOST client",
	Long:          `An IOST RPC client`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		iwalletSDK = sdk.NewIOSTDevSDK()
		iwalletSDK.SetChainID(chainID)
		iwalletSDK.SetServer(server)
		iwalletSDK.SetVerbose(verbose)
		iwalletSDK.SetSignAlgo(signAlgo)
		iwalletSDK.SetCheckResult(checkResult, checkResultDelay, checkResultMaxRetry)
		limit, err := ParseAmountLimit(amountLimit)
		if err != nil {
			return fmt.Errorf("invalid amount limit %v: %v", amountLimit, err)
		}
		iwalletSDK.SetTxInfo(gasLimit, gasRatio, expiration, delaySecond, limit)
		iwalletSDK.SetUseLongestChain(useLongestChain)
		return iwalletSDK.Connect()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		iwalletSDK.CloseConn()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	startTime := time.Now()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if verbose {
		fmt.Println("Executed in", time.Since(startTime))
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "configuration file (default $HOME/.iwallet.yaml)")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", true, "print verbose information")
	rootCmd.PersistentFlags().StringVarP(&accountName, "account", "a", "", "which account to use")
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "localhost:30002", "set server of this client")
	rootCmd.PersistentFlags().BoolVarP(&useLongestChain, "use_longest", "", false, "get info on longest chain")
	rootCmd.PersistentFlags().BoolVarP(&checkResult, "check_result", "", true, "check publish/call status after sending to chain")
	rootCmd.PersistentFlags().Float32VarP(&checkResultDelay, "check_result_delay", "", 3, "rpc checking will occur at [checkResultDelay] seconds after sending to chain.")
	rootCmd.PersistentFlags().Int32VarP(&checkResultMaxRetry, "check_result_max_retry", "", 20, "max times to call grpc to check tx status")
	rootCmd.PersistentFlags().StringVarP(&signAlgo, "sign_algo", "", "ed25519", "sign algorithm")
	rootCmd.PersistentFlags().Float64VarP(&gasLimit, "gas_limit", "l", 1000000, "gas limit for a transaction")
	rootCmd.PersistentFlags().Float64VarP(&gasRatio, "gas_ratio", "p", 1.0, "gas ratio for a transaction")
	rootCmd.PersistentFlags().StringVarP(&amountLimit, "amount_limit", "", "*:unlimited", "amount limit for one transaction, eg iost:300.00|ram:2000")
	rootCmd.PersistentFlags().Int64VarP(&expiration, "expiration", "e", 60*5, "expiration time for a transaction in seconds")
	rootCmd.PersistentFlags().Uint32VarP(&chainID, "chain_id", "", uint32(1024), "chain id which distinguishes different network")
	rootCmd.PersistentFlags().StringVarP(&txTime, "tx_time", "", "", "use the special tx time instead of now, format: 2019-01-22T17:00:39+08:00")

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
	server      string
	accountName string
	keyPair     *account.KeyPair
	signAlgo    string

	gasLimit    float64
	gasRatio    float64
	expiration  int64
	amountLimit string
	delaySecond int64

	checkResult         bool
	checkResultDelay    float32
	checkResultMaxRetry int32
	useLongestChain     bool

	verbose bool

	chainID uint32
)
