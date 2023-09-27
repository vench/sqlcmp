package sqlcmp

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Segment int

const (
	SegmentColumns Segment = 1 << iota
	SegmentFrom
	SegmentJoin
	SegmentWhere
	SegmentGroup
	SegmentOrder

	SegmentSkipValues

	SegmentAll = -1 ^ SegmentSkipValues
)

const delimiterSegment = "|"

var ErrParse = errors.New("parse error")

// SemiHash this function creates a hash of a request based on its segment.
func SemiHash(sql string, s Segment) (string, error) {
	p := NewParser(NewLexer(sql))

	stmt := p.parseSQLSelectStatement()

	if errs := p.Errors(); len(errs) != 0 {
		return "", fmt.Errorf("%w: %v", ErrParse, errs)
	}

	var sb strings.Builder

	if s&SegmentColumns != 0 {
		writeSegment(&sb, stmt.SQLSelectColumns, s)
	}

	if s&SegmentFrom != 0 {
		writeSegment(&sb, stmt.From, s)
	}

	if s&SegmentJoin != 0 {
		writeSegment(&sb, stmt.Join, s)
	}

	if s&SegmentWhere != 0 {
		writeSegment(&sb, stmt.Cond, s)
	}
	if s&SegmentGroup != 0 {
		writeSegment(&sb, stmt.Group, s)
	}
	if s&SegmentOrder != 0 {
		writeSegment(&sb, stmt.Order, s)
	}

	return hashString(sb.String())
}

func hashString(s string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(s)); err != nil {
		return "", err
	}

	sum := h.Sum(nil)

	return hex.EncodeToString(sum), nil
}

func writeSegment(b *strings.Builder, exp []Expression, s Segment) {
	if len(exp) == 0 {
		return
	}

	str := make([]string, len(exp))
	for i := range exp {
		if s&SegmentSkipValues != 0 {
			str[i] = structcher(exp[i])
		} else {
			str[i] = exp[i].String()
		}
	}

	sort.SliceStable(str, func(i, j int) bool {
		return str[i] > str[j]
	})

	for i := range str {
		b.WriteString(str[i])
		b.WriteString(delimiterSegment)
	}

	b.WriteString(delimiterSegment)
}
