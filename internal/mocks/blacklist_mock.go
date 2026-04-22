package mocks

import "time"

type MockBlacklist struct {
	IsBlacklistedFunc func(token string) (bool, error)
	AddFunc           func(token string, expiresAt time.Time) error
}

func (m *MockBlacklist) IsBlacklisted(token string) (bool, error) {
	if m.IsBlacklistedFunc != nil {
		return m.IsBlacklistedFunc(token)
	}
	return false, nil
}

func (m *MockBlacklist) Add(token string, expiresAt time.Time) error {
	if m.AddFunc != nil {
		return m.AddFunc(token, expiresAt)
	}
	return nil
}
