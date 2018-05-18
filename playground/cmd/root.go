package cmd

import (
	"fmt"
	"os"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/verifier"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var language string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "playground",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		db := Database{make(map[string][]byte)}
		mdb := state.NewDatabase(&db)
		pool := state.NewPool(mdb)
		for _, k := range viper.AllKeys() {
			v := viper.GetString(k)
			val, _ := state.ParseValue(v)
			pool.Put(state.Key(k), val)
		}

		v := verifier.NewCacheVerifier(pool)
		var sc0 vm.Contract

		switch language {
		case "lua":
			for i, file := range args {
				code := ReadSourceFile(file)
				parser, err := lua.NewDocCommentParser(code)
				if err != nil {
					panic(err)
				}
				sc, err := parser.Parse()
				if err != nil {
					panic(err)
				}
				if i == 0 {
					sc0 = sc
				}

				v.StartVM(sc)
			}
		default:
			fmt.Println(language, "not supported")
		}
		pool2, gas, err := v.Verify(sc0)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
		pool2.Flush()
		fmt.Println("======Report")
		fmt.Println("gas spend:", gas)
		fmt.Println("state trasition:")
		fmt.Println("state trasition:")
		for k, v := range db.Normal {
			fmt.Printf("  %v: %v\n", k, v)
		}

	},
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "values", "", "get values of test environment, default ./values.yaml")
	rootCmd.PersistentFlags().StringVarP(&language, "lang", "l", "lua", "set language of contract, default lua")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("./values.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
