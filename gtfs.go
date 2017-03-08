package main

import (
    "github.com/geops/gtfsparser"
    "github.com/geops/gtfsparser/gtfs"
    "fmt"
    "time"
)

func main() {

    // load feed
    start := time.Now()
    feed := gtfsparser.NewFeed()
    feed.Parse("/home/xianny/projects/rc/mta-tsp/data/google_transit.zip")
    elapsed := time.Since(start)
    fmt.Printf("Done, parsed %d agencies, %d stops, %d routes, %d trips, %d fare attributes\n\n",
        len(feed.Agencies), len(feed.Stops), len(feed.Routes), len(feed.Trips), len(feed.FareAttributes))
    fmt.Printf("Loading feed took %s\n", elapsed)


    services := map[string]bool {
        "A20161106WKD": true,
        "B20161106WKD": true,
        "R20161106WKD": true,
    }

    graph := map[*gtfs.Stop]map[*gtfs.StopTime][]*gtfs.StopTime {}

    
    start = time.Now()
    for _, trip := range feed.Trips {
        if services[trip.Service.Id] {
            for i, stoptime := range trip.StopTimes {
                if i+1 < len(trip.StopTimes) {
                    next_stoptime := trip.StopTimes[i+1]
                    if next_stoptime.Sequence - 1 == stoptime.Sequence {
                        connections := graph[stoptime.Stop]
                        connections[stoptime] = append(connections[stoptime], next_stoptime) // todo: insert in order?
                        graph[stoptime.Stop] = connections
                    }
                }
            }
        }
    }


    fmt.Printf("there are %d stops", len(graph))
    elapsed = time.Since(start)
    fmt.Printf("building edges took %s\n", elapsed)
    // fmt.Printf("%d stops")
    // fmt.Printf()


    // edges := []
    // for trip range feed.Trips
    //     edges.append
    // for k, v := range feed.Stops {
    //     fmt.Printf("[%s] %s (@ %f,%f)\n", k, v.Name, v.Lat, v.Lon)
    // }
}

func filterMap(l []feed.Trip, p func(feed.Trip) bool, f func(feed.Trip)) = {
    for _, trip := range l {
        if p(trip) {
            f(trip)
        }
    }
}