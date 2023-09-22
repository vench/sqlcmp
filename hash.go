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
		writeSegment(&sb, stmt.SQLSelectColumns)
	}

	if s&SegmentFrom != 0 {
		writeSegment(&sb, stmt.From)
	}

	if s&SegmentJoin != 0 {
		writeSegment(&sb, stmt.Join)
	}

	if s&SegmentWhere != 0 {
		// @todo: use SegmentSkipValues
		writeSegment(&sb, stmt.Cond)
	}
	if s&SegmentGroup != 0 {
		writeSegment(&sb, stmt.Group)
	}
	if s&SegmentOrder != 0 {
		writeSegment(&sb, stmt.Order)
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

func writeSegment(b *strings.Builder, exp []Expression) {
	if len(exp) == 0 {
		return
	}

	sort.SliceStable(exp, func(i, j int) bool {
		return exp[i].String() > exp[i].String()
	})

	for i := range exp {
		b.WriteString(exp[i].String())
		b.WriteString(delimiterSegment)
	}

	b.WriteString(delimiterSegment)
}
