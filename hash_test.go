package sqlcmp

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testHashString(t *testing.T, s string) string {
	t.Helper()

	h, err := hashString(s)
	require.NoError(t, err)

	return h
}

func TestSemiHash(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name    string
		sql     string
		out     string
		segment Segment
		err     error
	}{
		{
			name: "empty",
			out:  testHashString(t, ""),
		},
		{
			name: "error parse",
			sql:  "select * from wh as 1 ere id=1",
			err:  ErrParse,
		},
		{
			name:    "segment all",
			sql:     "select * from users where id = 100",
			segment: SegmentAll,
			out:     testHashString(t, "*||users||(id = 100)||"),
		},
		{
			name:    "segment from",
			sql:     "select * from users where id = 100",
			segment: SegmentFrom,
			out:     testHashString(t, "users||"),
		},
		{
			name:    "segment all and skip values",
			sql:     "select * from users where id = 100 and abc IN (99,100)",
			segment: SegmentAll | SegmentSkipValues,
			out:     testHashString(t, "*||users||(id = ?) ANDabc IN (?) AND||"),
		},
		{
			name:    "segment all and not skip values",
			sql:     "select * from users where id = 100 and abc IN (99,100)",
			segment: SegmentAll,
			out:     testHashString(t, "*||users||((id = 100) AND abc IN (99, 100))||"),
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, err := SemiHash(tc.sql, tc.segment)
			if tc.err == nil {
				require.NoError(t, err)
				require.EqualValues(t, tc.out, out)
			} else {
				require.Error(t, err)
				require.True(t, errors.Is(err, tc.err))
			}
		})
	}
}
