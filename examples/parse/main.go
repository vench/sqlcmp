package main

import (
	"fmt"
	"github.com/vench/sqlcmp"
	"os"
	"reflect"
)

func main() {
	input := `
SELECT teaser_id AS entity_id, SUM(shows) AS shows, SUM(clicks) AS clicks, SUM(goals) AS goals, SUM(currency_amount) AS amount 
FROM teasers 
WHERE (teaser_id IN (229, 999)) 
  AND (date='0000-00-00') 
GROUP BY teaser_id 
ORDER BY NULL`

	p := sqlcmp.NewParser(sqlcmp.NewLexer(input))
	s := p.ParseStatement()

	fmt.Printf("Print SQL query: \n%s\n", s.String())
	fmt.Printf("Print errors: %v\n", p.Errors())

	stmp, ok := s.(*sqlcmp.SQLSelectStatement)
	if !ok {
		fmt.Printf("failed cast stmp to SQLSelectStatement")
		os.Exit(1)
	}

	for i := range stmp.SQLSelectColumns {
		c := stmp.SQLSelectColumns[i]

		fmt.Printf("column: %s, typeof: %v\n", c.String(), reflect.TypeOf(c))
	}
}
