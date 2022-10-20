package libs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/montanaflynn/stats"
)

type Elastic struct {
	Logger       *Logger
	HttpClient   *http.Client
	HttpUsername string
	HttpPassword string
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (e *Elastic) runRequest(method string, url string, target interface{}, payload []byte) error {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		e.Logger.Instance.Fatal(err)
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(e.HttpUsername, e.HttpPassword))
	req.Header.Add("Content-Type", "application/json")

	resp, err := e.HttpClient.Do(req)
	if err != nil {
		e.Logger.Instance.Fatal(err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

type node struct {
	Name     string `json:"name"`
	DiskUsed int    `json:"disk.used,string"`
	NodeRole string `json:"node.role"`
}

type shard struct {
	Index string `json:"index"`
	Shard string `json:"shard"`
	State string `json:"state"`
	Store int    `json:"store,string"`
	Node  string `json:"node"`
}

type rebalanceInfo struct {
	Largest    node
	Smallest   node
	MeanSize   int
	DiffInSize struct {
		Largest  int
		Smallest int
	}
	DiffInSizePercent struct {
		Largest  int
		Smallest int
	}
	PercentOfDifference  int
	ProceedWithRebalance bool
}

type diskSpaceInfoResponse []node
type shardsInfoResponse []shard

type MoveCommand struct {
	Move struct {
		Index    string `json:"index"`
		Shard    string `json:"shard"`
		FromNode string `json:"from_node"`
		ToNode   string `json:"to_node"`
	} `json:"move"`
}

type ShardsReroute struct {
	Commands []MoveCommand `json:"commands"`
}

type rerouteResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

func findNodeByName(nodes []node, nodeName string) node {
	var result node
	for i := 1; i < len(nodes); i++ {
		if nodes[i].Name == nodeName {
			result = nodes[i]
			break
		}
	}

	return result
}

// GetDiskSpaceInfo func
func (e Elastic) GetDiskSpaceInfo(url string, allowedPercentOfDifference int, fromNode string, toNode string) rebalanceInfo {

	diskSpaceInfo := diskSpaceInfoResponse{}
	var payload []byte
	err := e.runRequest("GET", fmt.Sprintf("%s/_cat/nodes?v&h=name,disk.used,node.role&format=json&bytes=b", url), &diskSpaceInfo, payload)
	if err != nil {
		e.Logger.Instance.Fatal(err)
	}

	// filter only data nodes
	dataNodes := diskSpaceInfoResponse{}
	for i := range diskSpaceInfo {
		if strings.Contains(diskSpaceInfo[i].NodeRole, "d") {
			dataNodes = append(dataNodes, diskSpaceInfo[i])
		}
	}

	var nodeWithMinUsage node
	if toNode == "" {
		// finding node with min disk usage
		nodeWithMinUsage = dataNodes[0]
		for i := 1; i < len(dataNodes); i++ {
			if nodeWithMinUsage.DiskUsed > dataNodes[i].DiskUsed {
				nodeWithMinUsage = dataNodes[i]
			}
		}
	} else {
		nodeWithMinUsage = findNodeByName(dataNodes, toNode)
	}

	var nodeWithMaxUsage node
	if fromNode == "" {
		// finding node with max disk usage
		nodeWithMaxUsage = dataNodes[0]
		for i := 1; i < len(dataNodes); i++ {
			if nodeWithMaxUsage.DiskUsed < dataNodes[i].DiskUsed {
				nodeWithMaxUsage = dataNodes[i]
			}
		}
	} else {
		nodeWithMaxUsage = findNodeByName(dataNodes, fromNode)
	}

	// preparing data for stats package
	data := []float64{}
	for i := 0; i < len(dataNodes); i++ {
		data = append(data, float64(dataNodes[i].DiskUsed))
	}

	meanUsage, _ := stats.Mean(data)

	// diff in size between min/max and mean
	minDiffInSize := nodeWithMinUsage.DiskUsed - int(meanUsage)
	maxDiffInSize := nodeWithMaxUsage.DiskUsed - int(meanUsage)

	// percentage diff in size between min/max and mean
	minDiffInSizePercent := minDiffInSize * 100 / int(meanUsage)
	maxDiffInSizePercent := maxDiffInSize * 100 / int(meanUsage)

	// is rebalance needed?
	// proceedWithRebalance := maxDiffInSizePercent > int64(allowedPercentOfDifference)
	// e.Logger.Instance.Info(fmt.Sprintf("proceedWithRebalance = %+v", proceedWithRebalance))

	// # based on standard deviation percentage
	// # https://www.chem.tamu.edu/class/fyp/keeney/stddev.pdf
	standardDeviation, _ := stats.StandardDeviation(data)

	standardDeviationPercent := standardDeviation * 100 / meanUsage

	proceedWithRebalance := int(standardDeviationPercent) > allowedPercentOfDifference

	return rebalanceInfo{
		Largest:  nodeWithMaxUsage,
		Smallest: nodeWithMinUsage,
		MeanSize: int(meanUsage),
		DiffInSize: struct {
			Largest  int
			Smallest int
		}{
			Largest:  maxDiffInSize,
			Smallest: minDiffInSize,
		},
		DiffInSizePercent: struct {
			Largest  int
			Smallest int
		}{
			Largest:  maxDiffInSizePercent,
			Smallest: minDiffInSizePercent,
		},
		PercentOfDifference:  int(standardDeviationPercent),
		ProceedWithRebalance: proceedWithRebalance,
	}
}

func (e Elastic) GetShardsInfo(url string, diskSpaceInfo rebalanceInfo, shardsToMove int) []shard {
	// from/largest shards
	shardsInfo := shardsInfoResponse{}
	var payload []byte
	err := e.runRequest("GET", fmt.Sprintf("%s/_cat/shards?format=json&bytes=b", url), &shardsInfo, payload)
	if err != nil {
		e.Logger.Instance.Fatal(err)
	}

	// shards on source
	var shardsOnSource []shard
	for i := 0; i < len(shardsInfo); i++ {
		if shardsInfo[i].Node == diskSpaceInfo.Largest.Name {
			shardsOnSource = append(shardsOnSource, shardsInfo[i])
		}
	}

	// shards on target
	var shardsOnTarget []shard
	for i := 0; i < len(shardsInfo); i++ {
		if shardsInfo[i].Node == diskSpaceInfo.Smallest.Name {
			shardsOnTarget = append(shardsOnTarget, shardsInfo[i])
		}
	}

	var shardsAvailableForMove []shard
	// filtering out shards of indices already on target
	for i := 0; i < len(shardsOnSource); i++ {
		canBeMoved := true
		for j := 1; j < len(shardsOnTarget); j++ {
			if shardsOnSource[i].Index == shardsOnTarget[j].Index || shardsOnSource[i].State != "STARTED" {
				canBeMoved = false
				break
			}
		}

		if canBeMoved {
			shardsAvailableForMove = append(shardsAvailableForMove, shardsOnSource[i])
		}
	}

	// sorting by Store size, desc
	sort.Slice(shardsAvailableForMove, func(i, j int) bool {
		return shardsAvailableForMove[i].Store > shardsAvailableForMove[j].Store
	})

	// finding N largest shards
	shardsAvailableForMove = shardsAvailableForMove[:shardsToMove]

	return shardsAvailableForMove
}

func (e Elastic) PrepareMoveCommand(shards []shard, fromNode, toNode string) ShardsReroute {

	shardsReroute := ShardsReroute{}

	for i := 0; i < len(shards); i++ {

		command := MoveCommand{
			Move: struct {
				Index    string `json:"index"`
				Shard    string `json:"shard"`
				FromNode string `json:"from_node"`
				ToNode   string `json:"to_node"`
			}{
				Index:    shards[i].Index,
				Shard:    shards[i].Shard,
				FromNode: fromNode,
				ToNode:   toNode,
			},
		}

		shardsReroute.Commands = append(shardsReroute.Commands, command)

	}
	return shardsReroute
}

func (e Elastic) ExecuteMoveCommands(url string, moveCommands ShardsReroute, dryRun bool) rerouteResponse {

	bytesMoveCommand, _ := json.Marshal(moveCommands)
	// e.Logger.Instance.Info(fmt.Sprintf("moveCommands = %+v", string(bytesMoveCommand)))

	reroute := rerouteResponse{}
	err := e.runRequest("POST", fmt.Sprintf("%s/_cluster/reroute?dry_run="+strconv.FormatBool(dryRun), url), &reroute, bytesMoveCommand)
	if err != nil {
		e.Logger.Instance.Fatal(err)
	}

	return reroute
}
