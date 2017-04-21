package main

import (
    "github.com/geops/gtfsparser"
    "github.com/geops/gtfsparser/gtfs"
    "fmt"
    "time"
    "math"
)

const DATA_PATH = "./data/google_transit.zip"

type Graph map[string]Connections // station_parent_stop_id -> Connections
type Connections map[string]int // dest_station_parent_stop_id -> time to get there in seconds

func build_graph(feed *gtfsparser.Feed) (graph Graph) {
    services := map[string]bool {
        "A20161106WKD": true,
        "B20161106WKD": true,
        "R20161106WKD": true,
    }

    // map[stop.id, Connections]
    graph = make(Graph)

    start := time.Now()

    // travel between stations
    for _, trip := range feed.Trips {
        if services[trip.Service.Id] {
            for i, stoptime := range trip.StopTimes {
                if i+1 < len(trip.StopTimes) {
                    next_stoptime := trip.StopTimes[i+1]
                    if next_stoptime.Sequence - 1 == stoptime.Sequence {

                        connections := safe_get_connections(graph, stoptime.Stop.Parent_station)
                        next_stop := next_stoptime.Stop.Parent_station

                        distance, ok := connections[next_stop]
                        if !ok {
                            distance = seconds_between(stoptime.Departure_time, next_stoptime.Arrival_time)
                            connections[next_stop] = distance
                        }

                        graph[stoptime.Stop.Parent_station] = connections
                    }
                }
            }
        }
    }

    // transfers between lines at same station
    for _, transfer := range feed.Transfers {
        if transfer.From_stop != transfer.To_stop {
            connections := safe_get_connections(graph, transfer.From_stop.Id)
            _, ok := connections[transfer.To_stop.Id]
            if !ok {
                connections[transfer.To_stop.Id] = transfer.Min_transfer_time
            }
            graph[transfer.From_stop.Id] = connections
        }
    }

    elapsed := time.Since(start)
    fmt.Printf("building edges took %s\n", elapsed)
    return
}

func safe_get_connections(graph Graph, origin_station string) (connections Connections) {
    connections, ok := graph[origin_station]
        if !ok {
            connections = make(Connections)
        }
    graph[origin_station] = connections
    return connections
}


func extract_min_index(nodes []string, distances map[string]int) int {
    i := len(nodes)-1
    for j, node := range nodes {
        distance, ok := distances[node]
        if ok && distance < distances[nodes[i]] {
            i = j
        }
    }
    return i
}

func test_extract_min_index() bool {
    nodes := []string{"a", "b", "c", "d", "e"}
    distances := map[string]int{
        "a": 14,
        "c": 0,
        "e": 18,
    }
    return extract_min_index(nodes, distances) == 2
}

func remove(s []string, i int) []string {
    s[len(s) - 1], s[i] = s[i], s[len(s) - 1]
    return s[:len(s)-1]
}
func shortest_path(origin string, destination string, graph Graph) map[string]string {

    // use dijkstra's algorithm
    distances := make(map[string]int)
    previous := make(map[string]string)
    nodes := make([]string, len(graph))

    // initialization
    for origin_id, _ := range graph {
        distances[origin_id] = math.MaxUint32
        nodes = append(nodes, origin_id)
    }
    distances[origin] = 0


    for len(nodes) > 0 {

        i := extract_min_index(nodes, distances)
        current := nodes[i]

        if current == destination {
            break;
        }
        nodes = remove(nodes, i)

        for dest, time := range graph[current] {
            alt := distances[current] + time
            if alt < distances[dest] {
                distances[dest] = alt
                previous[dest] = current
            }
        }

    }

    return previous
}

func seconds_between(t1 string, t2 string) int {
    format := "15:04:05"
    _t1, _ := time.Parse(format, t1)
    _t2, _ := time.Parse(format, t2)
    return int(_t2.Sub(_t1)/time.Second)
}

func load_feed() (feed *gtfsparser.Feed) {
    start := time.Now()
    feed = gtfsparser.NewFeed()
    feed.Parse(DATA_PATH)
    elapsed := time.Since(start)
    fmt.Printf("Done, parsed %d agencies, %d stops, %d routes, %d trips, %d fare attributes\n\n",
        len(feed.Agencies), len(feed.Stops), len(feed.Routes), len(feed.Trips), len(feed.FareAttributes))
    fmt.Printf("Loading feed took %s\n", elapsed)
    return
}

func missing_stops(feed *gtfsparser.Feed, graph Graph) []*gtfs.Stop {
    missing_stops := []*gtfs.Stop{}
    for _, stop := range feed.Stops {
        last := stop.Id[len(stop.Id)-1:]
        _, ok := graph[stop.Id]
        if !ok && last != "N" && last != "S" {
            missing_stops = append(missing_stops, stop)
        }
    }
    return missing_stops
}

func main() {

    feed := load_feed()
    graph := build_graph(feed)


    origin := "L17" // myrtle-wyckoff
    destination := "R16" // times sq
    track := shortest_path(origin, destination, graph)

    curr := destination
    for track[curr] != origin {
        fmt.Printf("%s <-- ", curr)
        curr = track[curr]
    }
    fmt.Printf("\n")

}

// 92ms for map[stoptime -> list[stoptime]]
// 512ms
