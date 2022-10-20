package libs

import (
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func DryRunTable() {
	t := table.NewWriter()
	t.SetStyle(table.StyleDouble)
	t.Style().Color.Row = text.Colors{text.BgBlack, text.FgHiYellow}
	t.SetOutputMirror(os.Stdout)
	t.AppendRow(table.Row{"DRY-RUN: no operation will be performed"})
	t.Render()
}

func SettingsToTable(input map[string]interface{}) {
	var keys []interface{}
	var values []interface{}
	for k, v := range input {
		keys = append(keys, k)

		if strings.ToLower(k) == "password" {
			values = append(values, "--redacted--")
		} else {
			values = append(values, v)
		}
	}

	tableSettings := table.NewWriter()
	tableSettings.SetStyle(table.StyleDouble)
	tableSettings.SetOutputMirror(os.Stdout)
	tableSettings.SetTitle("Application settings")
	tableSettings.AppendHeader(keys)
	tableSettings.AppendRow(values)
	tableSettings.Render()
}

func RebalanceInfoToTable(input rebalanceInfo) {
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

	t := table.NewWriter()
	t.SetStyle(table.StyleDouble)
	t.Style().Options.SeparateRows = true
	t.SetTitle("Rebalance info")
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"Largest",
		"Largest",
		"Largest",
		"Smallest",
		"Smallest",
		"Smallest",
		"Mean Size",
		"Diff In Size",
		"Diff In Size",
		"Diff In Size Percent",
		"Diff In Size Percent",
		"Percent Of Difference",
		"Proceed With Rebalance",
	},
		rowConfigAutoMerge,
	)

	t.AppendHeader(table.Row{
		"Name",
		"Disk Used",
		"Node Role",
		"Name",
		"Disk Used",
		"Node Role",
		"",
		"Largest",
		"Smallest",
		"Largest",
		"Smallest",
	})

	t.AppendRow(table.Row{
		input.Largest.Name,
		input.Largest.DiskUsed,
		input.Largest.NodeRole,
		input.Smallest.Name,
		input.Smallest.DiskUsed,
		input.Smallest.NodeRole,
		input.MeanSize,
		input.DiffInSize.Largest,
		input.DiffInSize.Smallest,
		input.DiffInSizePercent.Largest,
		input.DiffInSizePercent.Smallest,
		input.PercentOfDifference,
		input.ProceedWithRebalance,
	},
		rowConfigAutoMerge,
	)
	t.Render()
}

func ShardsAvailableForMoveToTable(input []shard) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDouble)
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Shards available to move")
	t.AppendHeader(table.Row{
		"Node",
		"Index",
		"Shard",
		"State",
		"Store",
	})

	for _, v := range input {
		t.AppendRow(table.Row{
			v.Node,
			v.Index,
			v.Shard,
			v.State,
			v.Store,
		})
	}

	t.Render()
}

func MoveCommandsToTable(input []MoveCommand) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDouble)
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Shard move comamnds to execute")
	t.AppendHeader(table.Row{
		"From Node",
		"To Node",
		"Index",
		"Shard",
	})

	for _, v := range input {
		t.AppendRow(table.Row{
			v.Move.FromNode,
			v.Move.ToNode,
			v.Move.Index,
			v.Move.Shard,
		})
	}

	t.Render()
}

func RerouteResponseToTable(input rerouteResponse) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDouble)
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Shards reroute response")

	t.AppendRow(table.Row{
		"Acknowledged",
		input.Acknowledged,
	})

	t.Render()
}
