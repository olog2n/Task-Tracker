package mocks

import "database/sql"

type MockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func NewResult(lastInsertID, rowsAffected int64) *MockResult {
	return &MockResult{
		lastInsertID: lastInsertID,
		rowsAffected: rowsAffected,
	}
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.lastInsertID, nil
}

func (m *MockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

var _ sql.Result = (*MockResult)(nil)
