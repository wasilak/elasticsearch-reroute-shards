package libs

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

func Run() {
	elastic := Elastic{
		HttpClient:   &http.Client{Timeout: 10 * time.Second},
		HttpUsername: viper.GetString("username"),
		HttpPassword: viper.GetString("password"),
	}

	if viper.GetBool("dry-run") {
		if viper.GetString("log-format") == "json" {
			slog.Info("--------------- DRY-RUN no operation will be performed ---------------")
		} else {
			DryRunTable()
		}
	}

	rebalanceInfo := elastic.GetDiskSpaceInfo(viper.GetString("host"), viper.GetInt("allowed-percent-of-difference"), viper.GetString("from-node"), viper.GetString("to-node"))

	if viper.GetBool("debug") {
		if viper.GetString("log-format") == "json" {
			slog.Info(fmt.Sprintf("proceedWithRebalance = %+v", rebalanceInfo))
		} else {
			RebalanceInfoToTable(rebalanceInfo)
		}
	}

	shardsAvailableForMove := elastic.GetShardsInfo(viper.GetString("host"), rebalanceInfo, viper.GetInt("shards"))

	if viper.GetBool("debug") {
		if viper.GetString("log-format") == "json" {
			slog.Info(fmt.Sprintf("[%+v] shardsAvailableForMove = %+v", viper.GetInt("shards"), shardsAvailableForMove))
		} else {
			ShardsAvailableForMoveToTable(shardsAvailableForMove)
		}
	}

	moveCommands := elastic.PrepareMoveCommand(shardsAvailableForMove, rebalanceInfo.Largest.Name, rebalanceInfo.Smallest.Name)

	if viper.GetBool("debug") {
		if viper.GetString("log-format") == "json" {
			slog.Info(fmt.Sprintf("moveCommands = %+v", moveCommands))
		} else {
			MoveCommandsToTable(moveCommands.Commands)
		}
	}

	rerouteResponse := elastic.ExecuteMoveCommands(viper.GetString("host"), moveCommands, viper.GetBool("dry-run"))

	if viper.GetString("log-format") == "json" {
		slog.Info(fmt.Sprintf("rerouteResponse = %+v", rerouteResponse))
	} else {
		RerouteResponseToTable(rerouteResponse)
	}
}
