package handlers

import "github.com/stretchr/testify/mock"

type Logger interface {
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keyvals ...interface{}) {
	m.Called(msg, keyvals)
}

func (m *MockLogger) Error(msg string, keyvals ...interface{}) {
	m.Called(msg, keyvals)
}

func (m *MockLogger) Debug(msg string, keyvals ...interface{}) {
	m.Called(msg, keyvals)
}

func (m *MockLogger) Warn(msg string, keyvals ...interface{}) {
	m.Called(msg, keyvals)
}
