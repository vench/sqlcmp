package main

import (
	"bufio"
	"fmt"
	"github.com/vench/sqlcmp"
	"log"
	"os"
)

type queryKind struct {
	exampleQuery string
	count        int
}

// In this example of a grouped SQL query  by source (From section) and condition (Where section).
func main() {
	f, err := os.Open("./examples/aggregate_log/query.log")
	if err != nil {
		log.Fatalf("failed to open file: %v \n", err)
	}
	defer f.Close()

	mQueryKind := make(map[string]queryKind)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		query := scanner.Text()

		hash, err := sqlcmp.SemiHash(query, sqlcmp.SegmentWhere|sqlcmp.SegmentFrom|sqlcmp.SegmentSkipValues)
		if err != nil {
			log.Fatalf("failed to maske hase: %v from query: %s\n", err, query)
		}

		qk, ok := mQueryKind[hash]
		if !ok {
			qk = queryKind{
				exampleQuery: query, count: 0,
			}
		}

		qk.count++
		mQueryKind[hash] = qk
	}

	if err = scanner.Err(); err != nil {
		log.Fatalf("scanner error: %v \n", err)
	}

	for k := range mQueryKind {
		fmt.Printf("Query kind: %s, count: %d\n", mQueryKind[k].exampleQuery, mQueryKind[k].count)
	}
}
