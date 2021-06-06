package main

import (
	"strings"
)

type Query struct {
	Address string `json:"address"`
	EventLog string `json:"eventLog"`
	SelectStatements []string `json:"selectStatements"`
	LimitSize int `json:"limitSize"`
	FromBlockNumber float64 `json:"fromBlockNumber"`
	ToBlockNumber float64 `json:"toBlockNumber"`
	WhereClauses []WhereClause `json:"whereClauses"`
	Debug bool
}

type WhereClause struct {
	Name string `json:"name"`
	Value interface{} `json:"value"`
	ValueList []interface{}`json:"valueList"`
}

func NewQuery(Address string) Query {
	return Query{Address: Address}
}

func (q *Query) From(eventLog string) *Query {
	q.EventLog = eventLog
	q.FromBlockNumber = 1
	return q
}

func (q *Query) Select(selectStatement string) *Query {
	q.SelectStatements = append(
		q.SelectStatements,
		selectStatement)
	return q
}

func (q *Query) WhereIs(Name string, Value interface{}) *Query {
	where := WhereClause{
		Name: Name,
		Value: Value,
	}
	q.WhereClauses = append(
		q.WhereClauses,
		where)
	return q
}

func (q *Query) WhereIn(Name string, ValueList []interface{}) *Query {
	where := WhereClause{
		Name: Name,
		ValueList: ValueList,
	}
	q.WhereClauses = append(
		q.WhereClauses,
		where)
	return q
}

func (q *Query) ExecuteQuery(caches *Caches) []map[string]interface{} {
	cache := (*caches)[strings.ToLower(q.Address)]
	return queryForCache(cache, *q)
}
