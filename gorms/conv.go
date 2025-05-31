package gorms

import "gorm.io/gorm/clause"

type Column struct {
	Name string `json:"name"`
	Desc bool   `json:"desc"`
}

func (s *Column) Conv() clause.OrderByColumn {
	return clause.OrderByColumn{
		Column: clause.Column{Name: s.Name},
		Desc:   s.Desc,
	}
}

type Columns struct {
	Cols []Column `json:"cols"`
}
