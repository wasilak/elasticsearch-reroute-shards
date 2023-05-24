package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wasilak/elasticsearch-reroute-shards/libs"
	"golang.org/x/exp/slog"
)

var (
	rootCmd = &cobra.Command{
		Use:   libs.AppName,
		Short: "Elasticsearch shard rebalancing tool (based on size, not number of them per node)",
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetContext(ctx)
		},
		Run: rootFunc,
	}
	ctx = context.Background()
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(libs.InitConfig)

	rootCmd.PersistentFlags().StringVar(&libs.CfgFile, "config", "", "config file (default is $HOME/."+libs.AppName+"/config.yml)")
	rootCmd.PersistentFlags().BoolVar(&libs.CacheEnabled, "cacheEnabled", false, "cache enabled")
	rootCmd.PersistentFlags().StringVar(&libs.Listen, "listen", "127.0.0.1:3000", "listen address")

	rootCmd.PersistentFlags().String("from-node", "", "Node to move shards FROM")
	rootCmd.PersistentFlags().String("to-node", "", "Node to move shards TO")
	rootCmd.PersistentFlags().String("host", "http://localhost:9200", "Elasticsearch host address with port")
	rootCmd.PersistentFlags().Int("shards", 2, "Number of shards to move")
	rootCmd.PersistentFlags().Int("allowed-percent-of-difference", 10, "Allowed percent of difference in nodes disk used")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Perform dry-run, no changes will be applied to cluster")
	rootCmd.PersistentFlags().String("username", "", "Elasticsearch username")
	rootCmd.PersistentFlags().String("password", "", "Elasticsearch password")
	rootCmd.PersistentFlags().String("log-format", "plain", "Log format [json, plain]")
	rootCmd.PersistentFlags().String("log-level", "INFO", "Log level")

	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("cacheEnabled", rootCmd.PersistentFlags().Lookup("cacheEnabled"))

	viper.BindPFlag("from-node", rootCmd.PersistentFlags().Lookup("from-node"))
	viper.BindPFlag("to-node", rootCmd.PersistentFlags().Lookup("to-node"))
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("shards", rootCmd.PersistentFlags().Lookup("shards"))
	viper.BindPFlag("allowed-percent-of-difference", rootCmd.PersistentFlags().Lookup("allowed-percent-of-difference"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))

	rootCmd.AddCommand(versionCmd)
}

func rootFunc(cmd *cobra.Command, args []string) {
	libs.InitLogging(viper.GetString("log-level"), viper.GetString("log-format"))

	slog.Debug(fmt.Sprintf("%+v", viper.AllSettings()))

	if viper.GetBool("log-level") {
		if viper.GetString("log-format") == "json" {
			slog.Debug("all settings", slog.AnyValue(viper.AllSettings()))
		} else {
			libs.SettingsToTable(viper.AllSettings())
		}
	}

	libs.Run()
}
