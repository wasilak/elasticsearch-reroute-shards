package libs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var log string

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url string, target interface{}) error {
    r, err := myClient.Get(url)
    if err != nil {
        return err
    }
    defer r.Body.Close()

    return json.NewDecoder(r.Body).Decode(target)
}

type node struct {
	Name string `json:"name"`
	DiskUsedString string `json:"disk.used"`
	DiskUsed int64
	NodeRole string `json:"node.role"`
}

type diskSpaceInfoResponse []node

// GetDiskSpaceInfo func
func GetDiskSpaceInfo(log string, url string) {

	diskSpaceInfo := diskSpaceInfoResponse{}
    getJSON(fmt.Sprintf("%s/_cat/nodes?v&h=name,disk.used,node.role&format=json&bytes=b", url), &diskSpaceInfo)

	// strconv.ParseInt(diskSpaceInfo.DiskUsedString, 10, 64)
	for i := range diskSpaceInfo {
		var err error
		diskSpaceInfo[i].DiskUsed, err = strconv.ParseInt(diskSpaceInfo[i].DiskUsedString, 10, 64)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		// fmt.Printf("%+v\n", node.DiskUsedString)

	}
	

	fmt.Printf("%+v\n", fmt.Sprintf("%s/_cat/nodes?v&h=name,disk.used,node.role&format=json&bytes=b", url))
	fmt.Printf("%+v\n", diskSpaceInfo)

    // df_space = pd.DataFrame.from_dict(disk_space_info, orient='columns')
    // df_space['disk.used'] = pd.to_numeric(df_space['disk.used'])

    // df_space = df_space[df_space['node.role'].str.contains("d")]

    // largest = df_space.nlargest(1, "disk.used")
    // smallest = df_space.nsmallest(1, "disk.used")

    // mean_size = df_space["disk.used"].mean()

    // largest = largest.to_dict(orient='records')[0]
    // smallest = smallest.to_dict(orient='records')[0]

    // largest_diff_in_size = largest["disk.used"] - mean_size
    // smallest_diff_in_size = smallest["disk.used"] - mean_size

    // largest_diff_in_size_percent = largest_diff_in_size * 100 / mean_size
    // smallest_diff_in_size_percent = smallest_diff_in_size * 100 / mean_size

    // # manually calculated percentage
    // # proceed_with_rebalance = largest_diff_in_size_percent > allowed_percent_of_difference

    // # based on standard deviation percentage
    // # https://www.chem.tamu.edu/class/fyp/keeney/stddev.pdf
    // standard_deviation_percent = round(df_space.std()["disk.used"] * 100 / df_space.mean()["disk.used"], 1)
    // proceed_with_rebalance = standard_deviation_percent > arguments.allowed_percent_of_difference

    // return proceed_with_rebalance, {
    //     "largest": largest,
    //     "smallest": smallest,
    //     "mean_size": mean_size,
    //     "diff_in_size": {
    //         "largest": largest_diff_in_size,
    //         "smallest": smallest_diff_in_size,
    //     },
    //     "diff_in_size_percent": {
    //         "largest": largest_diff_in_size_percent,
    //         "smallest": smallest_diff_in_size_percent,
    //     },
    //     "percent_of_difference": standard_deviation_percent,
    // }
}
