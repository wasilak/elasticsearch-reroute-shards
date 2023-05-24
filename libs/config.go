package libs

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

const AppName = "elasticsearch-reroute-shards"
const AppVersion = "v1.0.0"

var (
	CfgFile      string
	Listen       string
	CacheEnabled bool
)

func InitConfig() {
	godotenv.Load()

	viper.SetEnvPrefix(AppName)

	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.SetConfigType("yaml")
		viper.SetConfigName(AppName)
		viper.AddConfigPath(home)
		viper.AddConfigPath("./")
	}
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Printf("%+v\n", err)
	}

	if strings.ToLower(viper.GetString("loglevel")) == "debug" {
		log.Println(slog.AnyValue(viper.AllSettings()))
	}
}
