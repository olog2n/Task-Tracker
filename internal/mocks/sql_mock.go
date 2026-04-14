package mocks

import "database/sql"

type Result struct {
	lastInsertID int64
	rowsAffected int64
}

func NewResult(lastInsertID, rowsAffected int64) *Result {
	return &Result{
		lastInsertID: lastInsertID,
		rowsAffected: rowsAffected,
	}
}

func (m *Result) LastInsertId() (int64, error) {
	return m.lastInsertID, nil
}

func (m *Result) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

var _ sql.Result = (*Result)(nil)
