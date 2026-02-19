package main

import (
	"bfm/fzf/algo"
	"bfm/fzf/util"
	"strings"
)

type termType int

const (
	termFuzzy termType = iota
	termExact
	termExactBoundary
	termPrefix
	termSuffix
	termEqual
)

type Term struct {
	Pattern string
	Type    termType
	Inverse bool
}

type Query struct {
	Terms []Term
	Or    bool // true if OR logic, false for AND
}

func ParseQuery(query string) *Query {
	q := &Query{}
	if strings.Contains(query, "|") {
		q.Or = true
		parts := strings.Split(query, "|")
		for _, part := range parts {
			q.Terms = append(q.Terms, parseTerm(strings.TrimSpace(part)))
		}
	} else {
		q.Or = false
		parts := strings.Fields(query)
		for _, part := range parts {
			q.Terms = append(q.Terms, parseTerm(part))
		}
	}
	return q
}

func parseTerm(s string) Term {
	typ := termFuzzy
	inv := false
	if strings.HasPrefix(s, "!") {
		inv = true
		typ = termExact
		s = s[1:]
	}
	if s != "$" && strings.HasSuffix(s, "$") {
		typ = termSuffix
		s = s[:len(s)-1]
	}
	if len(s) > 2 && strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		typ = termExactBoundary
		s = s[1 : len(s)-1]
	} else if strings.HasPrefix(s, "'") {
		typ = termExact
		s = s[1:]
	}
	if strings.HasPrefix(s, "^") {
		if typ == termSuffix {
			typ = termEqual
		} else {
			typ = termPrefix
		}
		s = s[1:]
	}
	return Term{Pattern: s, Type: typ, Inverse: inv}
}

func (q *Query) Eval(text string) int {
	chars := util.RunesToChars([]rune(text))
	score := 0
	for _, term := range q.Terms {
		termScore := q.evalTerm(term, &chars)
		if q.Or {
			if termScore > 0 {
				score = termScore
				break
			}
		} else {
			if termScore == 0 {
				return 0
			}
			if termScore > score {
				score = termScore
			}
		}
	}
	if q.Or && score == 0 {
		return 0
	}
	return score
}

func (q *Query) evalTerm(term Term, chars *util.Chars) int {
	if term.Pattern == "" {
		return 1
	}
	pattern := []rune(term.Pattern)
	var res algo.Result
	var score int
	switch term.Type {
	case termExact:
		res, _ = algo.ExactMatchNaive(false, true, true, chars, pattern, false, nil)
		score = res.Score
	case termExactBoundary:
		res, _ = algo.ExactMatchBoundary(false, true, true, chars, pattern, false, nil)
		score = res.Score
	case termPrefix:
		res, _ = algo.PrefixMatch(false, true, true, chars, pattern, false, nil)
		score = res.Score
	case termSuffix:
		res, _ = algo.SuffixMatch(false, true, true, chars, pattern, false, nil)
		score = res.Score
	case termEqual:
		res, _ = algo.EqualMatch(false, true, true, chars, pattern, false, nil)
		score = res.Score
	default: // termFuzzy
		res, _ = algo.FuzzyMatchV2(false, true, true, chars, pattern, false, nil)
		score = res.Score
	}
	if term.Inverse {
		if score > 0 {
			return 0
		}
		return 1
	}
	return score
}
