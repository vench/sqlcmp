# SQL compare

This package allows you to parse SQL queries into tokens and compare them in parts.

### Examples 

Comparing two SQL queries by FROM and WHERE sections:

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
 
