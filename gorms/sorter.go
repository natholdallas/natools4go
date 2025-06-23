package gorms

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Sorter struct {
	Name string `json:"name"`
	Desc bool   `json:"desc"`
}

func (s *Sorter) Conv() clause.OrderByColumn {
	return clause.OrderByColumn{Column: clause.Column{Name: s.Name}, Desc: s.Desc}
}

func (s *Sorter) Sort(tx *gorm.DB) {
	tx.Order(s.Conv())
}

type Sorters struct {
	Cols []Sorter `query:"cols" json:"cols"`
}

func (s *Sorters) Conv() []clause.OrderByColumn {
	res := []clause.OrderByColumn{}
	for i := range s.Cols {
		res = append(res, s.Cols[i].Conv())
	}
	return res
}

func (s *Sorters) Sort(tx *gorm.DB) {
	tx.Order(clause.OrderBy{Columns: s.Conv()})
}
