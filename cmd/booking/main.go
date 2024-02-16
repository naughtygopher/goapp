package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	adtechTeams := map[string]struct{}{
		"081-084": {},
		"085-092": {},
		"092-100": {},
		"153-156": {},
		"161-168": {},
		"177-184": {},
	}

	firstFloorBTeamBookings := []string{
		"076-104",
		"081-084",
		"085-092",
		"088-096",
		"092-100",
		"093-104",
		"104",
		"153-156",
		"153-189",
		"156-164",
		"161-168",
		"161-168",
		"177-184",
		"177-184 ",
		"177-192",
		"185-192",
		"197-184",
	}

	for _, srange := range firstFloorBTeamBookings {
		parts := strings.Split(srange, "-")
		start, _ := strconv.Atoi(parts[0])
		end := -1
		if len(parts) == 2 {
			end, _ = strconv.Atoi(parts[1])
		}
		if end > start {
			fmt.Printf("range: %d - %d\n", start, end)
			for start <= end {
				fmt.Printf("%03d\n", start)
				start++
			}
		} else {
			fmt.Println(start)
		}
	}

	_ = firstFloorBTeamBookings
	_ = adtechTeams
}
