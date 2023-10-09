package cmd

import (
	"fmt"
	"os"

	"github.com/fabriziosestito/gostubpkg/pkg/gen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "gostubpkg",
	Short: "gostubpkg is a tool to create fake packages",
	// TODO: add long description
	Long: "todo: add long description",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputDir := viper.GetString("output-dir")
		generateGoMod := viper.GetBool("generate-go-mod")
		allowImports := viper.GetStringSlice("allow-import")
		functionsBodies := viper.GetStringMapString("function-body")

		err := gen.GenerateStubs(args, outputDir, generateGoMod, allowImports, functionsBodies)
		if err != nil {
			cobra.CheckErr(err)
		}
	},
}

// Execute executes the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var (
		outputDir       string
		generateGoMod   bool
		allowImports    []string
		functionsBodies map[string]string
	)
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/gostubpkg.yaml)")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Specify the output directory for the stubs. (default is $PWD)")
	rootCmd.Flags().BoolVarP(&generateGoMod, "generate-go-mod", "m", false, "Generate the go.mod file in the root of the stub package.")
	rootCmd.Flags().StringArrayVarP(&allowImports, "allow-import", "a", nil, "Specify this flag multiple times to add external imports\nthat will not be removed from the generated stubs.\nExample: -a k8s.io/api/core/v1")
	rootCmd.Flags().StringToStringVarP(&functionsBodies, "function-body", "f", nil, "Specify this flag multiple times to add a type mapping.\nExample: -f cmd.Execute='println(\"hello world\")' -f yourpkg.(*YourType).YourMethod='return nil'")

	err := viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		cobra.CheckErr(err)
	}
}

// initConfig reads in config file if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("gostubpkg.yaml")
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
