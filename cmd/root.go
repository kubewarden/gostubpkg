package cmd

import (
	"fmt"
	"os"

	"github.com/fabriziosestito/go-stub-package/pkg/gen"
	"github.com/spf13/cobra"
)

var (
	generateGoMod   bool
	allowImports    []string
	functionsBodies map[string]string
)

var rootCmd = &cobra.Command{
	Use:   "go-stub-package",
	Short: "go-stub-package is a tool to create fake packages",
	// TODO: add long description
	Long: "todo: add long description",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := gen.GenerateStubs(args, generateGoMod, allowImports, functionsBodies)
		if err != nil {
			panic(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&generateGoMod, "generate-go-mod", "m", true, "Generate the go.mod file in the root of the stub package.")
	rootCmd.Flags().StringArrayVarP(&allowImports, "allow-import", "a", nil, "Specify this flag multiple times to add external imports\nthat will not be removed from the generated stubs.\nExample: -a k8s.io/api/core/v1")
	rootCmd.Flags().StringToStringVarP(&functionsBodies, "function-body", "f", nil, "Specify this flag multiple times to add a type mapping.\nExample: -f cmd.Execute='println(\"hello world\")' -f yourpkg.(*YourType).YourMethod='return nil'")
}
