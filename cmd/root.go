package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/fabriziosestito/gostubpkg/pkg/gen"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	k       = koanf.New(".")
)

var rootCmd = &cobra.Command{
	Use:   "gostubpkg",
	Short: "gostubpkg is a tool to create fake packages",
	// TODO: add long description
	Long: "todo: add long description",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil)
		if err != nil {
			log.Fatalf("error loading flags: %v", err)
		}

		outputDir := k.String("output-dir")
		generateGoMod := k.Bool("generate-go-mod")
		functionBodies := k.StringMap("function-bodies")
		allowImports := k.Strings("allow-imports")

		err = gen.GenerateStubs(args, outputDir, generateGoMod, allowImports, functionBodies)
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
		outputDir      string
		generateGoMod  bool
		allowImports   []string
		functionBodies map[string]string
	)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "gostubpkg.yaml", "config file (default is $PWD/gostubpkg.yaml)")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Specify the output directory for the stubs. (default is $PWD)")
	rootCmd.Flags().BoolVarP(&generateGoMod, "generate-go-mod", "m", false, "Generate the go.mod file in the root of the stub package.")
	rootCmd.Flags().StringSliceVarP(&allowImports, "allow-imports", "a", nil, "Specify this flag multiple times to add external imports\nthat will not be removed from the generated stubs.\nExample: -a k8s.io/api/core/v1")
	rootCmd.Flags().StringToStringVarP(&functionBodies, "function-bodies", "f", nil, "Specify this flag multiple times to add a type mapping.\nExample: -f \"cmd.Execute\"='println(\"hello world\")' -f \"yourpkg.(*YourType).YourMethod\"='return nil'")
}

// initConfig reads in config file if set.
func initConfig() {
	err := k.Load(file.Provider(cfgFile), yaml.Parser())
	if err == nil {
		log.Printf("using config file: %s", cfgFile)
	} else if !os.IsNotExist(err) {
		cobra.CheckErr(fmt.Errorf("error loading config file: %v", err))
	}
}
