package iwallet

import (
	"fmt"
	"os"

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
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "configuration file (default $HOME/.iwallet.yaml)")

	rootCmd.PersistentFlags().BoolVarP(&sdk.verbose, "verbose", "", true, "print verbose information")
	rootCmd.PersistentFlags().StringVarP(&sdk.accountName, "account", "", "", "which account to use")
	rootCmd.PersistentFlags().StringVarP(&sdk.server, "server", "s", "localhost:30002", "set server of this client")
	rootCmd.PersistentFlags().BoolVarP(&sdk.useLongestChain, "use_longest", "", false, "get info on longest chain")
	rootCmd.PersistentFlags().BoolVarP(&sdk.checkResult, "check_result", "", true, "check publish/call status after sending to chain")
	rootCmd.PersistentFlags().Float32VarP(&sdk.checkResultDelay, "check_result_delay", "", 3, "rpc checking will occur at [checkResultDelay] seconds after sending to chain.")
	rootCmd.PersistentFlags().Int32VarP(&sdk.checkResultMaxRetry, "check_result_max_retry", "", 20, "max times to call grpc to check tx status")
	rootCmd.PersistentFlags().StringVarP(&sdk.signAlgo, "sign_algo", "", "ed25519", "sign algorithm")
	rootCmd.PersistentFlags().Float64VarP(&sdk.gasLimit, "gas_limit", "l", 1000000, "gas limit for a transaction")
	rootCmd.PersistentFlags().Float64VarP(&sdk.gasRatio, "gas_ratio", "p", 1.0, "gas ratio for a transaction")
	rootCmd.PersistentFlags().StringVarP(&sdk.amountLimit, "amount_limit", "", "*:unlimited", "amount limit for one transaction, eg iost:300.00|ram:2000")
	rootCmd.PersistentFlags().Int64VarP(&sdk.expiration, "expiration", "e", 60*5, "expiration time for a transaction in seconds")
	rootCmd.PersistentFlags().Uint32VarP(&sdk.chainID, "chain_id", "", uint32(1024), "chain id which distinguishes different network")
	rootCmd.PersistentFlags().StringVarP(&sdk.txTime, "tx_time", "", "", "use the special tx time instead of now, format: 2019-01-22T17:00:39+08:00")

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
