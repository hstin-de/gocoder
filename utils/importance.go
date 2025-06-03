package utils

import (
	"compress/gzip"
	"encoding/csv"
	"hstin/gocoder/config"
	"io"
	"log"
	"os"
	"strconv"
)

var ImportanceMap map[string]int

func LoadImportanceMap() {

	f, err := os.Open(config.WikimediaImportance)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		log.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gz.Close()

	reader := csv.NewReader(gz)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	ImportanceMap := make(map[string]int)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(record) > 3 {
			importanceFloat, _ := strconv.ParseFloat(record[3], 64)
			ImportanceMap[record[4]] = int(importanceFloat * config.WikimediaMaxImportance)
		}
	}
}
