package utils

import (
	"hstin/gocoder/mapping"
)

func CreateRank(tags map[string]string, pop int64) int {

	// Calculate the rank based on the place type
	calculatedRank := mapping.PlaceRank[tags["place"]]

	// Factor in population
	if pop < 50000000 {
		calculatedRank += 750
	} else if pop < 10000000 {
		calculatedRank += 500
	} else if pop < 1000000 {
		calculatedRank += 200
	} else if pop < 500000 {
		calculatedRank += 150
	} else if pop < 100000 {
		calculatedRank += 100
	} else if pop < 50000 {
		calculatedRank += 75
	} else if pop < 10000 {
		calculatedRank += 50
	} else if pop < 5000 {
		calculatedRank += 25
	} else if pop < 1000 {
		calculatedRank += 10
	} else if pop < 500 {
		calculatedRank += 5
	} else if pop < 100 {
		calculatedRank += 1
	}

	// Capital cities get a bonus
	if tags["capital"] == "yes" {
		calculatedRank += 500
	}

	calculatedRank += ImportanceMap[tags["wikidata"]]

	return calculatedRank
}
