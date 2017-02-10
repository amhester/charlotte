package query_parser

import (
	"bytes"
	"errors"
	"sync"

	QueryModels "github.com/amhester/charlotte/protos"
)

type work struct {
	Error     error
	QueryPart *QueryModels.QueryPart
	PartIdx   int
}

func ParseQuery(query string) (QueryModels.Query, error) {
	var err error
	raw := []byte(query)
	var resQuery = QueryModels.Query{}
	wg := &sync.WaitGroup{}
	c := make(chan work)
	escaped := false
	leftR := false
	partIdx := 0
	lastIdx := 0
	for i := 0; i < len(raw); i++ {
		char := raw[i]
		switch char {
		case 34: // -> \"
			escaped = !escaped
			continue
		case 45: // -> -
			if escaped || (partIdx%2 == 1 && !leftR) {
				continue
			}
			leftR = false
		case 60: // -> <
			if escaped {
				continue
			}
			leftR = true
		case 62: // -> >
			if escaped || (i != 0 && raw[i-1] == 61) {
				continue
			}
		case 61: // -> =
			if escaped {
				continue
			}
		default:
			continue
		}
		wg.Add(1)
		go buildQueryPart(c, wg, raw[lastIdx:i], partIdx)
		partIdx++
		lastIdx = i
	}
	resQuery.QueryParts = make([]*QueryModels.QueryPart, partIdx+1)
	go monitorWork(wg, c)
	for w := range c {
		if w.Error != nil {
			err = w.Error
		}
		resQuery.QueryParts[w.PartIdx] = w.QueryPart
	}
	return resQuery, err
}

func buildQueryPart(c chan work, wg *sync.WaitGroup, raw []byte, idx int) {
	w := work{PartIdx: idx}
	defer wg.Done()
	q := &QueryModels.QueryPart{}
	switch raw[0] {
	case 45: // -> -
		q.Type = QueryModels.QueryPartType_Edge
		raw = raw[1:]
	case 60: // -> <
		q.Type = QueryModels.QueryPartType_Edge
		raw = raw[2:]
	case 61: // -> =
		q.Type = QueryModels.QueryPartType_Output
		raw = raw[2:]
	default:
		q.Type = QueryModels.QueryPartType_Node
	}
	escaped := false
	lastIdx := 0
	for i := 0; i < len(raw); i++ {
		if w.Error != nil {
			break
		}
		char := raw[i]
		if char == 34 { // -> \"
			escaped = !escaped
			continue
		}
		if char == 40 { // -> (
			if !escaped {
				q.EntityType = raw[lastIdx:i]
				lastIdx = i
			}
			continue
		}
		if char == 41 { // -> )
			if !escaped {
				q.Filters, w.Error = buildFilterExpression(raw[lastIdx+1 : i])
				lastIdx = i
			}
			continue
		}
		if char == 91 { // -> [
			if !escaped {
				lastIdx = i
			}
			continue
		}
		if char == 93 {
			if !escaped { // -> ]
				q.Captured, w.Error = buildCaptureExpression(raw[lastIdx+1 : i])
			}
			continue
		}
	}
	w.QueryPart = q
	c <- w
}

func buildCaptureExpression(raw []byte) (*QueryModels.DataCapture, error) {
	var err error
	length := len(raw)
	if length == 0 {
		return nil, nil
	}
	dc := &QueryModels.DataCapture{}
	n := bytes.Count(raw, []byte{44}) + 1
	fields := make([][]byte, n)
	fieldIdx := 0
	escaped := false
	lastIdx := 0
	for i := 0; i < length; i++ {
		char := raw[i]
		if char == 34 { // -> \"
			escaped = !escaped
			continue
		}
		if char == 32 { // -> " "
			///TODO: Maybe incrementing lastIdx here would fix issues with capturing trash white-space (at least for leading white-space between tokens)
			continue
		}
		if char == 58 { // -> :
			if !escaped {
				dc.VarName = raw[lastIdx:i]
				lastIdx = i + 1
			}
			continue
		}
		if char == 44 { // -> ,
			if !escaped {
				fields[fieldIdx] = raw[lastIdx:i]
				lastIdx = i + 1
				fieldIdx++
			}
			continue
		}
		if i+1 == length {
			if escaped {
				err = errors.New("Syntax Error: Unexpected end of capture expression.")
				continue
			}
			cap := raw[lastIdx : i+1]
			if len(cap) == 0 {
				err = errors.New("SyntaxError: Unexpected end of capture expression.")
				continue
			}
			fields[fieldIdx] = cap
			continue
		}
	}
	return dc, err
}

func buildFilterExpression(raw []byte) ([]*QueryModels.Filter, error) {
	length := len(raw)
	if length == 0 {
		return []*QueryModels.Filter{}, nil
	}
	var err error
	n := bytes.Count(raw, []byte{44}) + 1
	filters := make([]*QueryModels.Filter, n)
	///TODO: Fix to be like it is for buildCaptureExpression func
	j := 0
	lastIdx := 0
	for i := 0; i < n; i++ {
		filters[i] = &QueryModels.Filter{}
		escaped := false
		for ; j < length; j++ {
			char := raw[j]
			if char == 34 { // -> \"
				escaped = !escaped
				continue
			}
			if char == 32 { // -> " "
				///TODO: Maybe incrementing lastIdx here would fix issues with capturing trash white-space (at least for leading and trailing white-space between tokens)
				continue
			}
			if char == 58 { // -> :
				if !escaped {
					filters[i].Field = raw[lastIdx:j]
					lastIdx = j + 1
				}
				continue
			}
			if char == 44 { // -> ,
				if !escaped {
					filters[i].Value = raw[lastIdx:j]
					lastIdx = j + 1
					break
				}
			}
			if j+1 == length {
				if escaped {
					err = errors.New("Syntax Error: Unexpected end of capture expression.")
					continue
				}
				cap := raw[lastIdx : j+1]
				if len(cap) == 0 {
					err = errors.New("SyntaxError: Unexpected end of capture expression.")
					continue
				}
				filters[i].Value = []byte(cap)
				continue
			}
		}
	}
	///TODO: maybe perform some sane checks on generated filters, such as whether or not the last filter is actually there (if not, indicates unnecessary trailing comma in filter expression -> syntaxError)
	return filters, err
}

func monitorWork(wg *sync.WaitGroup, c chan work) {
	wg.Wait()
	close(c)
}
