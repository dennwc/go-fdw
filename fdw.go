package main

import "fmt"

func init() {
	SetTable(helloTable{Rows: 10})
}

type helloTable struct{ Rows int }

func (t helloTable) Stats() TableStats {
	return TableStats{Rows: uint(t.Rows), StartCost: 10, TotalCost: 1000}
}
func (t helloTable) Scan(rel *Relation, opts map[string]string) Iterator {
	return &helloIter{t: t, rel: rel}
}

var _ Explainable = (*helloIter)(nil)

type helloIter struct {
	t   helloTable
	rel *Relation
	row int
}

func (it *helloIter) Explain(e Explainer) {
	e.Property("Powered by", "Go FDW")
}
func (it *helloIter) Next() []interface{} {
	if it.row >= it.t.Rows {
		return nil
	}
	out := make([]interface{}, len(it.rel.Attr.Attrs))
	for i := range out {
		attr := it.rel.Attr.Attrs[i]
		if !attr.NotNull {
			continue
		}
		switch attr.Type {
		case TypeInt16, TypeInt32, TypeInt64:
			out[i] = int(it.row)
		case TypeText:
			out[i] = fmt.Sprintf("Row: %d, Col: %q", it.row, attr.Name)
		}
	}
	it.row++
	return out
}
func (it *helloIter) Reset()       { it.row = 0 }
func (it *helloIter) Close() error { return nil }
