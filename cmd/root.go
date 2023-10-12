package cmd

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/kubewarden/gostubpkg/pkg/gen"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	k       = koanf.New(".")
)

var rootCmd = &cobra.Command{
	Use:   "gostubpkg [flags] <patterns>...",
	Short: "gostubpkg is a tool for generating stubs of Go packages.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, patterns []string) {
		err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil)
		if err != nil {
			log.Fatalf("error loading flags: %v", err)
		}

		logrus.SetLevel(logrus.Level(int(logrus.InfoLevel) + k.Int("verbose")))

		inputDir := k.String("input-dir")
		outputDir := k.String("output-dir")
		generateGoMod := k.Bool("generate-go-mod")
		functionBodies := k.StringMap("function-bodies")
		allowImports := k.Strings("allow-imports")

		err = gen.GenerateStubs(inputDir, patterns, outputDir, generateGoMod, allowImports, functionBodies)
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
		inputDir       string
		outputDir      string
		generateGoMod  bool
		allowImports   []string
		functionBodies map[string]string
		verbose        int
	)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "gostubpkg.yaml", "config file")
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "Increase output verbosity. Example: --verbose=2 or -vv")

	rootCmd.Flags().StringVarP(&inputDir, "input-dir", "i", "", "Specify the directory in which to run the build system's query tool that provides information about the packages (default $PWD)")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Specify the output directory for the stubs (default $PWD)")
	rootCmd.Flags().BoolVarP(&generateGoMod, "generate-go-mod", "m", false, "Generate the go.mod file in the root of the stub package")
	rootCmd.Flags().StringSliceVarP(&allowImports, "allow-imports", "a", nil, "Specify this flag multiple times to add external imports\nthat will not be removed from the generated stubs.\nExample: -a k8s.io/api/core/v1")
	rootCmd.Flags().StringToStringVarP(&functionBodies, "function-bodies", "f", nil, "Specify this flag multiple times to add a type mapping.\nExample: -f \"cmd.Execute\"='println(\"hello world\")' -f \"yourpkg.(*YourType).YourMethod\"='return nil'")
}

// initConfig reads in config file if set.
func initConfig() {
	err := k.Load(file.Provider(cfgFile), yaml.Parser())
	if err == nil {
		log.Infof("using config file: %s", cfgFile)
	} else if !os.IsNotExist(err) {
		cobra.CheckErr(fmt.Errorf("error loading config file: %v", err))
	}
}
