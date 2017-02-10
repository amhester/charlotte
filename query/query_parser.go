package query_parser

import (
	"bytes"
	"errors"
	"sync"

	QueryModels "github.com/amhester/charlotte/protos"
)

func ParseQuery(query string) (*QueryModels.QueryPart, error) {
	var err error
	raw := []byte(query)
	var resQuery = &QueryModels.QueryPart{}
	wg := &sync.WaitGroup{}
	escaped := false
	leftR := false
	partIdx := 0
	lastIdx := 0
	previous := resQuery
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
		newPart := &QueryModels.QueryPart{}
		previous.Next = newPart
		previous = newPart
		wg.Add(1)
		go buildQueryPart(wg, raw[lastIdx:i], newPart)
		partIdx++
		lastIdx = i
	}
	wg.Wait()
	return resQuery, err
}

func buildQueryPart(wg *sync.WaitGroup, raw []byte, q *QueryModels.QueryPart) {
	var err error
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
		if err != nil {
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
				q.Filters, err = buildFilterExpression(raw[lastIdx+1 : i])
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
				q.Captured, err = buildCaptureExpression(raw[lastIdx+1 : i])
			}
			continue
		}
	}
	wg.Done()
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
