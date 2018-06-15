package cmd

import (
	"fmt"
	"os"

	"strings"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/verifier"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var language string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "playground",
	Short: "Playground of IOST script",
	Long: `Playground of IOST script, usage:
	playground a.lua b.lua ... --values value.yml
Playground runs lua script by turns.
`,
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
			if strings.Contains(k, ".") {
				kf := strings.Split(k, ".")
				fmt.Println(kf)
				pool.PutHM(state.Key(kf[0]), state.Key(kf[1]), val)
			} else {
				pool.Put(state.Key(k), val)
			}
		}

		v := verifier.NewCacheVerifier()
		//var sc0 vm.contract

		var (
			pool2 state.Pool
			gas   uint64
		)

		pool2 = pool.Copy()

		switch language {
		case "lua":
			for _, file := range args {
				code := ReadSourceFile(file)
				parser, err := lua.NewDocCommentParser(code)
				if err != nil {
					panic(err)
				}
				parser.Debug = true
				sc, err := parser.Parse()
				if err != nil {
					panic(err)
				}

				sc.SetPrefix(file[strings.LastIndex(file, "/")+1 : strings.LastIndex(file, ".")])

				v.StartVM(sc)

				pool2, gas, err = v.Verify(sc, pool2)
				if err != nil {
					fmt.Println("error:", err.Error())
				}
			}
		default:
			fmt.Println(language, "not supported")
		}

		pool2.Flush()
		fmt.Println("======Report")
		fmt.Println("gas spend:", gas)
		fmt.Println("state trasition:")
		for k, v := range db.Normal {
			fmt.Printf("  %v: %v\n", k, string(v))
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
