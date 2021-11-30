import requests
import json
import argparse
import pandas as pd
import logging

auth = ("", "")


def prepare_move_command(item, from_node, to_node):
    command = {
        "move": {
            "index": item.loc['index'],
            "shard": item.loc['shard'],
            "from_node": from_node,
            "to_node": to_node,
        }
    }
    return command


def get_shards_info(arguments, disk_space_info):
    # from/largest shards
    shards = requests.get("{}/_cat/shards?format=json&bytes=b".format(arguments.host), auth=auth).json()

    df_from = pd.DataFrame.from_dict(shards, orient='columns')

    df_from['store'] = pd.to_numeric(df_from['store'])

    largest = df_from[df_from['node'].str.contains(arguments.from_node) & ~df_from['index'].str.startswith(".") & df_from['state'].str.contains("STARTED")]

    largest_already_on_target = df_from[df_from['node'].str.contains(arguments.to_node) & ~df_from['index'].str.startswith(".") & df_from['state'].str.contains("STARTED")]

    # calculating difference in order to not try move shards belonging to indices already on target node
    largest = largest[~largest["index"].isin(largest_already_on_target["index"])]

    # smallest = df_from[df_from['node'].str.contains(arguments.to_node) & ~df_from['index'].str.startswith(".") & ~df_from['node'].str.contains(arguments.from_node)]
    # smallest = smallest.nsmallest(arguments.shards, "store")

    # semi-smart shards number calculation
    if arguments.shards == 0:
        sum = 0
        shards_number = 0

        while sum <= disk_space_info["diff_in_size"]["largest"]:
            shards_number = shards_number + 1
            shards = largest.nlargest(shards_number, "store")
            sum = shards.sum()["store"]

        shards = largest.nlargest(shards_number - 1, "store")
        sum = shards.sum()["store"]

    else:
        shards = largest.nlargest(arguments.shards, "store")

    return shards


def get_disk_space_info(arguments):

    disk_space_info = requests.get("{}/_cat/nodes?v&h=name,disk.used,node.role&format=json&bytes=b".format(arguments.host), auth=auth).json()

    df_space = pd.DataFrame.from_dict(disk_space_info, orient='columns')
    df_space['disk.used'] = pd.to_numeric(df_space['disk.used'])

    df_space = df_space[df_space['node.role'].str.contains("d")]

    largest = df_space.nlargest(1, "disk.used")
    smallest = df_space.nsmallest(1, "disk.used")

    mean_size = df_space["disk.used"].mean()

    largest = largest.to_dict(orient='records')[0]
    smallest = smallest.to_dict(orient='records')[0]

    largest_diff_in_size = largest["disk.used"] - mean_size
    smallest_diff_in_size = smallest["disk.used"] - mean_size

    largest_diff_in_size_percent = largest_diff_in_size * 100 / mean_size
    smallest_diff_in_size_percent = smallest_diff_in_size * 100 / mean_size

    # manually calculated percentage
    # proceed_with_rebalance = largest_diff_in_size_percent > allowed_percent_of_difference

    # based on standard deviation percentage
    # https://www.chem.tamu.edu/class/fyp/keeney/stddev.pdf
    standard_deviation_percent = round(df_space.std()["disk.used"] * 100 / df_space.mean()["disk.used"], 1)
    proceed_with_rebalance = standard_deviation_percent > arguments.allowed_percent_of_difference

    return proceed_with_rebalance, {
        "largest": largest,
        "smallest": smallest,
        "mean_size": mean_size,
        "diff_in_size": {
            "largest": largest_diff_in_size,
            "smallest": smallest_diff_in_size,
        },
        "diff_in_size_percent": {
            "largest": largest_diff_in_size_percent,
            "smallest": smallest_diff_in_size_percent,
        },
        "percent_of_difference": standard_deviation_percent,
    }


def execute_move_commands(arguments, move_commands):
    shards_reroute = {
        "commands": move_commands.to_dict(orient='records')
    }

    logging.info("Move JSON simulation: {}".format(json.dumps(shards_reroute)))

    if arguments.dry_run:
        logging.info("DRY_RUN!!!")
        return

    reroute_headers = {
        "Content-Type": "application/json"
    }

    reroute_result = requests.post("{}/_cluster/reroute".format(arguments.host), data=json.dumps(shards_reroute), headers=reroute_headers, auth=auth)

    reroute_result = reroute_result.json()

    # response cleanup
    if "state" in reroute_result:
        del(reroute_result["state"])

    logging.info("move result: {}".format(reroute_result))


def main():

    parser = argparse.ArgumentParser(
        description="Elasticsearch shard rebalancing tool (based on size, not number of them per node)",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )

    parser.add_argument('--from-node', type=str, help="Node to move shards FROM", default="")
    parser.add_argument('--to-node', type=str, help="Node to move shards TO", default="")
    parser.add_argument('--host', type=str, default="http://localhost:9200", help="Elasticsearch host address with port")
    parser.add_argument('--shards', "-s", type=int, help="Number of shards to move", default=0)
    parser.add_argument('--allowed-percent-of-difference', type=int, help="Allowed percent of difference in nodes disk used", default=10)
    parser.add_argument('--dry-run', dest="dry_run", action="store_true", help="Perform dry-run, no changes will be applied to cluster")
    parser.add_argument('--debug', dest="debug", action="store_true", help="Debug logging")

    arguments = parser.parse_args()

    logging.basicConfig(format='%(asctime)s %(levelname)s %(message)s', level=logging.DEBUG if arguments.debug else logging.INFO)

    proceed_with_rebalance, disk_space_info = get_disk_space_info(arguments)

    if not proceed_with_rebalance:
        logging.info("There is no need for rebalancing, percent of difference in nodes disk used is {} <= {} :)".format(disk_space_info["percent_of_difference"], arguments.allowed_percent_of_difference))
        exit(0)
    else:
        logging.info("Proceeding wirh rebalancing, percent of difference in nodes disk used is {} > {} :)".format(disk_space_info["percent_of_difference"], arguments.allowed_percent_of_difference))

    if len(arguments.from_node) == 0:
        arguments.from_node = disk_space_info["largest"]["name"]

    if len(arguments.to_node) == 0:
        arguments.to_node = disk_space_info["smallest"]["name"]

    logging.info("Moving shards: {} => {}".format(arguments.from_node, arguments.to_node))

    shards_to_move = get_shards_info(arguments, disk_space_info)

    logging.info(shards_to_move)

    move_commands_from = shards_to_move.apply(prepare_move_command, axis=1, result_type="expand", from_node=arguments.from_node, to_node=arguments.to_node)
    # move_commands_to = smallest.apply(prepare_move_command, axis=1, result_type="expand", from_node=to_node, to_node=from_node)
    # move_commands = move_commands_from.append(move_commands_to)

    move_commands = move_commands_from

    execute_move_commands(arguments, move_commands)


if __name__ == "__main__":
    main()
