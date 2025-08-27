package gorms

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Sorter struct {
	Column string `query:"column" json:"column"`
	Desc   bool   `query:"desc" json:"desc"`
}

func (s *Sorter) Scope(tx *gorm.DB) *gorm.DB {
	if s.Column != "" {
		return tx.Order(s.Conv())
	}
	return tx
}

func (s *Sorter) Conv() clause.OrderByColumn {
	return clause.OrderByColumn{Column: clause.Column{Name: s.Column}, Desc: s.Desc}
}

func (s *Sorter) Sort(tx *gorm.DB) {
	if s.Column != "" {
		tx.Order(s.Conv())
	}
}

type Sorters struct {
	Columns []Sorter `query:"columns" json:"columns"`
}

func (s *Sorters) Scope(tx *gorm.DB) *gorm.DB {
	if len(s.Columns) > 0 {
		return tx.Order(clause.OrderBy{Columns: s.Conv()})
	}
	return tx
}

func (s *Sorters) Conv() []clause.OrderByColumn {
	res := []clause.OrderByColumn{}
	for i := range s.Columns {
		res = append(res, s.Columns[i].Conv())
	}
	return res
}

func (s *Sorters) Sort(tx *gorm.DB) {
	if len(s.Columns) > 0 {
		tx.Order(clause.OrderBy{Columns: s.Conv()})
	}
}
