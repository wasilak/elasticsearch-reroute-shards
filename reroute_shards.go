package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/wasilak/es_rebalance/libs"
)

// Params type
type Params struct {
    FromNode *string
    ToNode *string
    Host *string
    Shards *int
    AllowedPercentOfDifference *int
    DryRun *bool
    Debug  *bool
}

func main(){
    file, err := os.OpenFile("./reroute_shards.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
		log.Fatal(err)
	}
    mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFormatter(&log.JSONFormatter{})

    log.SetFormatter(&log.JSONFormatter{})
    
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Elasticsearch shard rebalancing tool (based on size, not number of them per node)\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [params]\n", os.Args[0])
		flag.PrintDefaults()
    }

    params := Params{}
    
    params.FromNode = flag.String("from-node", "", "Node to move shards FROM")
    params.ToNode = flag.String("to-node", "", "Node to move shards TO")
    params.Host = flag.String("host", "http://localhost:9200", "Elasticsearch host address with port")
    params.Shards = flag.Int("shards", 2, "Number of shards to move")
    params.AllowedPercentOfDifference = flag.Int("allowed-percent-of-difference", 10, "Allowed percent of difference in nodes disk used")
    params.DryRun = flag.Bool("dry-run", false, "Perform dry-run, no changes will be applied to cluster")
    params.Debug = flag.Bool("debug", false, "Debug logging")

    flag.Parse()

    log.Info(*(params.FromNode))

    libs.GetDiskSpaceInfo(log, *(params.Host))
    
}
