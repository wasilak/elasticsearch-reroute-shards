package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wasilak/elasticsearch-reroute-shards/libs"
)

const AppVersion = "0.0.5"

func main() {

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Elasticsearch shard rebalancing tool (based on size, not number of them per node)\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [params]\n", os.Args[0])
		pflag.PrintDefaults()
	}

	flag.String("from-node", "", "Node to move shards FROM")
	flag.String("to-node", "", "Node to move shards TO")
	flag.String("host", "http://localhost:9200", "Elasticsearch host address with port")
	flag.Int("shards", 2, "Number of shards to move")
	flag.Int("allowed-percent-of-difference", 10, "Allowed percent of difference in nodes disk used")
	flag.Bool("dry-run", false, "Perform dry-run, no changes will be applied to cluster")
	flag.Bool("debug", false, "Debug logging")
	flag.String("username", "", "Elasticsearch username")
	flag.String("password", "", "Elasticsearch password")
	flag.String("log-format", "plain", "Log format [json, plain]")
	flag.String("log-file", "", "Log file path")
	flag.Bool("version", false, "Print current version")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("SRR")
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	if viper.GetBool("version") {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	logger := new(libs.Logger)

	logger.Init(viper.GetString("log-format"), viper.GetString("log-file"))

	if viper.GetBool("debug") == true {
		logger.Instance.Info(viper.AllSettings())
	}

	elastic := libs.Elastic{
		Logger:       logger,
		HttpClient:   &http.Client{Timeout: 10 * time.Second},
		HttpUsername: viper.GetString("username"),
		HttpPassword: viper.GetString("password"),
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if viper.GetBool("dry-run") {
		// logger.Instance.Info("--------------- DRY-RUN no operation will be performed ---------------")
		t1 := table.NewWriter()
		t1.SetStyle(table.StyleDouble)
		t1.SetOutputMirror(os.Stdout)
		t1.AppendRow(table.Row{"DRY-RUN no operation will be performed"})
		t1.Render()
	}

	t.AppendHeader(table.Row{"#", "First Name", "Last Name", "Salary"})
	t.AppendRows([]table.Row{
		{1, "Arya", "Stark", 3000},
		{20, "Jon", "Snow", 2000, "You know nothing, Jon Snow!"},
	})
	t.AppendSeparator()
	t.AppendRow([]interface{}{300, "Tyrion", "Lannister", 5000})
	t.AppendFooter(table.Row{"", "", "Total", 10000})
	t.Render()

	rebalanceInfo := elastic.GetDiskSpaceInfo(viper.GetString("host"), viper.GetInt("allowed-percent-of-difference"), viper.GetString("from-node"), viper.GetString("to-node"))

	if viper.GetBool("debug") {
		logger.Instance.Info(fmt.Sprintf("proceedWithRebalance = %+v", rebalanceInfo))
	}

	shardsAvailableForMove := elastic.GetShardsInfo(viper.GetString("host"), rebalanceInfo, viper.GetInt("shards"))

	if viper.GetBool("debug") {
		logger.Instance.Info(fmt.Sprintf("[%+v] shardsAvailableForMove = %+v", viper.GetInt("shards"), shardsAvailableForMove))
	}

	moveCommands := elastic.PrepareMoveCommand(shardsAvailableForMove, rebalanceInfo.Largest.Name, rebalanceInfo.Smallest.Name)

	if viper.GetBool("debug") {
		logger.Instance.Info(fmt.Sprintf("moveCommands = %+v", moveCommands))
	}

	elastic.ExecuteMoveCommands(viper.GetString("host"), moveCommands, viper.GetBool("dry-run"))

}
