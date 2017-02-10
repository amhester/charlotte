package messages_query

import (
	"fmt"
)

func (qp *QueryPart) ToString() string {
	if qp == nil {
		return "nil"
	}
	return fmt.Sprintf("{ Type: %s, EntityType: %s, Filters: %s, Captured: %s, Output: %s }", QueryPartType_name[int32(qp.Type)], string(qp.EntityType), stringifyFilters(qp.Filters), qp.Captured.ToString(), qp.Output.ToString())
}

func stringifyFilters(filters []*Filter) string {
	s := "["
	for _, f := range filters {
		s = fmt.Sprintf("%s, %s", s, f.ToString())
	}
	s = s + "]"
	return s
}

func (f *Filter) ToString() string {
	if f == nil {
		return "nil"
	}
	return fmt.Sprintf("{ Field: %s, Value: %s }", string(f.Field), string(f.Value))
}

func (dc *DataCapture) ToString() string {
	if dc == nil {
		return "nil"
	}
	s := fmt.Sprintf("{ VarName: %s, Fields: [", string(dc.VarName))
	for f := range dc.Fields {
		s = fmt.Sprintf("%s, %s", s, string(f))
	}
	s = s + "]}"
	return s
}

func (o *DataStructure) ToString() string {
	if o == nil {
		return "nil"
	}
	s := "["
	for _, p := range o.Props {
		s = fmt.Sprintf("%s, %s", s, p.ToString())
	}
	s = s + "]"
	return s
}

func (p *OutputProp) ToString() string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("{ Key: %s, Nested: %s }", string(p.Key), p.Nested.ToString())
}
