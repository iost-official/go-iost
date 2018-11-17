package iwallet

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iwallet",
	Short: "IOST client",
	Long:  `An IOST RPC client`,
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iwallet.yaml)")

	rootCmd.PersistentFlags().StringVarP(&sdk.accountName, "account", "", "", "which account to use")
	rootCmd.PersistentFlags().StringVarP(&sdk.server, "server", "s", "localhost:30002", "Set server of this client")
	rootCmd.PersistentFlags().BoolVarP(&sdk.checkResult, "checkResult", "", true, "Check publish/call status after sending to chain")
	rootCmd.PersistentFlags().Float32VarP(&sdk.checkResultDelay, "checkResultDelay", "", 3, "RPC checking will occur at [checkResultDelay] seconds after sending to chain.")
	rootCmd.PersistentFlags().Int32VarP(&sdk.checkResultMaxRetry, "checkResultMaxRetry", "", 10, "Max times to call grpc to check tx status")
	rootCmd.PersistentFlags().StringVarP(&sdk.signAlgo, "signAlgo", "", "ed25519", "Sign algorithm")
	rootCmd.PersistentFlags().Int64VarP(&sdk.gasLimit, "gaslimit", "l", 10000, "gasLimit for a transaction")
	rootCmd.PersistentFlags().Int64VarP(&sdk.gasPrice, "gasprice", "p", 100, "gasPrice for a transaction")
	rootCmd.PersistentFlags().StringVarP(&sdk.amountLimit, "amountLimit", "", "", "amount limit for one transaction, eg iost:300.00|ram:2000")
	rootCmd.PersistentFlags().Int64VarP(&sdk.expiration, "expiration", "e", 60*5, "expiration time for a transaction,for example,-e 60 means the tx will expire after 60 seconds from now on")

	//rootCmd.PersistentFlags().StringVarP(&dest, "dest", "d", "default", "Set destination of output file")
	//rootCmd.Flags().StringSliceVarP(&signers, "signers", "n", []string{}, "signers who should sign this transaction")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
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
