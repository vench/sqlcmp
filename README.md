# SQL compare

`sqlcmp` is a Go package designed to facilitate the comparison of SQL queries by parsing them into tokens and allowing for the comparison of these tokens in parts. This functionality is particularly useful for applications that need to determine if two SQL queries are semantically equivalent, ignoring differences in formatting, whitespace, or the order of clauses.

The package provides a function SemiHash that takes a SQL query and a mask specifying which parts of the query to consider for comparison. The mask can include segments such as `WHERE`, `FROM`, and `SkipValues`, which allows for a flexible comparison of queries. For example, two queries with the same `WHERE` and `FROM` clauses but in different orders, or with different values in the WHERE clause, can be considered equivalent if the mask is set to compare these segments.

### Examples 

Comparing two SQL queries by `FROM` and `WHERE` sections:

```go

import ( 
"fmt"
"github.com/vench/sqlcmp" 
)

var (
    query1 = "SELECT * FROM phones WHERE  date > '2023-09-27' AND user_id = 1004;"
    query2 = "SELECT * FROM phones WHERE  user_id = 1004 AND date > '2023-09-27';"
    mask = sqlcmp.SegmentWhere|sqlcmp.SegmentFrom|sqlcmp.SegmentSkipValues
)

hash1, _ := sqlcmp.SemiHash(query1, mask)
hash2, _ := sqlcmp.SemiHash(query2, mask)

fmt.Println(hash1 == hash2) // print true

```

You can find more examples in the directory [examples](./examples).

### TODO list
 
