package main

import (
    "github.com/geops/gtfsparser"
    // "github.com/geops/gtfsparser/gtfs"
    "github.com/golang/protobuf/proto"
    "fmt"
    "time"
    "math"
    "io/ioutil"
    "net/http"
)

const API_KEY = "c9a09f8f24aeb6a4e167cb9b82fa8eff"

type Connections struct {
    origin string //parent stop id
    trips map[string]int //destination_parent_stop_id, time to get there
}

func fetch_rt_feed(url string) (*FeedMessage, error) {
    fm := new(FeedMessage)
    res, err := http.Get(url)
    if err != nil {
        return fm, err
    }
    defer res.Body.Close()
    byteArray, err := ioutil.ReadAll(res.Body)

    err = proto.Unmarshal(byteArray, fm)
    if err != nil {
        return new(FeedMessage), err
    }
    return fm, err
}

func build_graph(feed *gtfsparser.Feed) (graph map[string]Connections) {
    services := map[string]bool {
        "A20161106WKD": true,
        "B20161106WKD": true,
        "R20161106WKD": true,
    }

    // map[stop.id, Connections]
    graph = make(map[string]Connections)
    
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

                        distance, ok := connections.trips[next_stop]
                        if !ok {
                            distance = seconds_between(stoptime.Departure_time, next_stoptime.Arrival_time)
                            connections.trips[next_stop] = distance
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
            _, ok := connections.trips[transfer.To_stop.Id]
            if !ok {
                connections.trips[transfer.To_stop.Id] = transfer.Min_transfer_time
            }
            graph[transfer.From_stop.Id] = connections
        }
    }

    elapsed := time.Since(start)
    fmt.Printf("building edges took %s\n", elapsed)
    return
}

func safe_get_connections(graph map[string]Connections, origin_station string) (connections Connections) {
    connections, ok := graph[origin_station]
        if !ok {
            connections = Connections{origin_station, make(map[string]int)}
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
func shortest_path(origin string, destination string, graph map[string]Connections) map[string]string {
    
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

        fmt.Printf("current node is %s, time is %d\n", current, distances[current])
        if current == destination {
            break;
        }
        nodes = remove(nodes, i)
        fmt.Printf("%d nodes left\n", len(nodes))

        for dest, time := range graph[current].trips {
            alt := distances[current] + time
            if alt < distances[dest] {
                distances[dest] = alt
                previous[dest] = current
            }
        }

    }

    return previous
}

// hh:mm:ss format
func seconds_between(t1 string, t2 string) int {
    layout := "15:04:05"
    _t1, _ := time.Parse(layout, t1)
    _t2, _ := time.Parse(layout, t2)
    return int(_t2.Sub(_t1)/time.Second)
}


func main() {

    rt_feed, _ := fetch_rt_feed(fmt.Sprintf("http://datamine.mta.info/mta_esi.php?key=%s&feed_id=1", API_KEY))
    fmt.Printf("Found %d entities in realtime feed\n", len(rt_feed.GetEntity()))

    for _, entity := range rt_feed.GetEntity() {
        if entity.GetTripUpdate() != nil {
            fmt.Printf("Trip update: %v\n", *entity.GetTripUpdate())
        }
    }
//     // load feed
//     start := time.Now()
//     feed := gtfsparser.NewFeed()
//     feed.Parse("/home/xianny/projects/rc/mta-tsp/data/google_transit.zip")
//     elapsed := time.Since(start)
//     fmt.Printf("Done, parsed %d agencies, %d stops, %d routes, %d trips, %d fare attributes\n\n",
//         len(feed.Agencies), len(feed.Stops), len(feed.Routes), len(feed.Trips), len(feed.FareAttributes))
//     fmt.Printf("Loading feed took %s\n", elapsed)

//     graph := build_graph(feed)


//     missing_stops := []*gtfs.Stop{}
//     for _, stop := range feed.Stops {
//         last := stop.Id[len(stop.Id)-1:]
//         _, ok := graph[stop.Id]
//         if !ok && last != "N" && last != "S" {
//             missing_stops = append(missing_stops, stop)
//             fmt.Printf("%s\n", stop.Id)
//         }
//     }

//     parent_stops := []*gtfs.Stop{}
//     child_stops := []*gtfs.Stop{}
//     for _, stop := range feed.Stops {
//         last := stop.Id[len(stop.Id)-1:]
//         if (last == "N" || last == "S") {
//             child_stops = append(child_stops, stop)
//         } else {
//             parent_stops = append(parent_stops, stop)
//         }
//     }

//     fmt.Println("there are %d stops", len(graph))

//     fmt.Println("%d missing stops", len(missing_stops))
//     fmt.Println("%d child stops", len(child_stops))
//     fmt.Println("%d parent stops", len(parent_stops))


//     origin := "L17" // myrtle-wyckoff
//     destination := "R16" // times sq
//     track := shortest_path(origin, destination, graph)

//     curr := destination
//     for track[curr] != origin {
//         fmt.Printf("%s <-- ", curr)
//         curr = track[curr]
//     }

//     fmt.Printf("test: %v\n", test_extract_min_index())

//     // test := "S10"
//     // fmt.Printf("trips from %s, %v", test, graph[test].trips)

//     tests := []string{"L17", "L16", "L15", "L14", "L13", "L12", "L11", "L10", "L08"}
//     for _, t := range tests {
//         fmt.Printf("trips from %s: [%v]", t, graph[t].trips)
//     }
//     // fmt.Printf("%v <- ", destination)
//     // curr := track[destination]
//     // for curr != origin {
//     //     fmt.Printf("%v <- ", curr)
//     //     curr = track[curr]
//     // }
    
//     // trip := shortest_path("M08", "M20", graph)
//     // fmt.Printf("%d length trip", len(trip))
}

// 92ms for map[stoptime -> list[stoptime]]
// 512ms 

    // fmt.Println("there are %d stops", len(graph))

    // fmt.Println("%d missing stops", len(missing_stops))
    // fmt.Println("%d child stops", len(child_stops))
    // fmt.Println("%d parent stops", len(parent_stops))

    // missing_stops := []*gtfs.Stop{}
    // for _, stop := range feed.Stops {
    //     last := stop.Id[len(stop.Id)-1:]
    //     if last != "N" && last != "S" && graph[stop.Id] == nil {
    //         missing_stops = append(missing_stops, stop)
    //         fmt.Printf("%s\n", stop.Id)
    //     }
    // }

    // parent_stops := []*gtfs.Stop{}
    // child_stops := []*gtfs.Stop{}
    // for _, stop := range feed.Stops {
    //     last := stop.Id[len(stop.Id)-1:]
    //     if (last == "N" || last == "S") {
    //         child_stops = append(child_stops, stop)
    //     } else {
    //         parent_stops = append(parent_stops, stop)
    //     }
    // }