package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/wasilak/elasticsearch-reroute-shards/libs"
)

var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of " + libs.AppName,
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.SetContext(ctx)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := versionFunc(); err != nil {
			return err
		}
		return nil
	},
}

func versionFunc() error {
	buildInfo, _ := debug.ReadBuildInfo()
	fmt.Printf("%s\nVersion %s (GO %s)\n", libs.AppName, version, buildInfo.GoVersion)
	return nil
}
