package gorms

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Sorter struct {
	Name string `query:"name" json:"name"`
	Desc bool   `query:"desc" json:"desc"`
}

func (s *Sorter) Conv() clause.OrderByColumn {
	return clause.OrderByColumn{Column: clause.Column{Name: s.Name}, Desc: s.Desc}
}

func (s *Sorter) Sort(tx *gorm.DB) {
	if s.Name != "" {
		tx.Order(s.Conv())
	}
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
	if len(s.Cols) > 0 {
		tx.Order(clause.OrderBy{Columns: s.Conv()})
	}
}
