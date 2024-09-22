// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/eac0de/getmetrics/internal/api/router (interfaces: IDatabase,IMetricsStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/eac0de/getmetrics/internal/models"
	gomock "github.com/golang/mock/gomock"
)

// MockIDatabase is a mock of IDatabase interface.
type MockIDatabase struct {
	ctrl     *gomock.Controller
	recorder *MockIDatabaseMockRecorder
}

// MockIDatabaseMockRecorder is the mock recorder for MockIDatabase.
type MockIDatabaseMockRecorder struct {
	mock *MockIDatabase
}

// NewMockIDatabase creates a new mock instance.
func NewMockIDatabase(ctrl *gomock.Controller) *MockIDatabase {
	mock := &MockIDatabase{ctrl: ctrl}
	mock.recorder = &MockIDatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIDatabase) EXPECT() *MockIDatabaseMockRecorder {
	return m.recorder
}

// PingContext mocks base method.
func (m *MockIDatabase) PingContext(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingContext", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PingContext indicates an expected call of PingContext.
func (mr *MockIDatabaseMockRecorder) PingContext(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingContext", reflect.TypeOf((*MockIDatabase)(nil).PingContext), arg0)
}

// MockIMetricsStore is a mock of IMetricsStore interface.
type MockIMetricsStore struct {
	ctrl     *gomock.Controller
	recorder *MockIMetricsStoreMockRecorder
}

// MockIMetricsStoreMockRecorder is the mock recorder for MockIMetricsStore.
type MockIMetricsStoreMockRecorder struct {
	mock *MockIMetricsStore
}

// NewMockIMetricsStore creates a new mock instance.
func NewMockIMetricsStore(ctrl *gomock.Controller) *MockIMetricsStore {
	mock := &MockIMetricsStore{ctrl: ctrl}
	mock.recorder = &MockIMetricsStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIMetricsStore) EXPECT() *MockIMetricsStoreMockRecorder {
	return m.recorder
}

// GetMetric mocks base method.
func (m *MockIMetricsStore) GetMetric(arg0 context.Context, arg1, arg2 string) (*models.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetric", arg0, arg1, arg2)
	ret0, _ := ret[0].(*models.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetric indicates an expected call of GetMetric.
func (mr *MockIMetricsStoreMockRecorder) GetMetric(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetric", reflect.TypeOf((*MockIMetricsStore)(nil).GetMetric), arg0, arg1, arg2)
}

// ListAllMetrics mocks base method.
func (m *MockIMetricsStore) ListAllMetrics(arg0 context.Context) ([]*models.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAllMetrics", arg0)
	ret0, _ := ret[0].([]*models.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAllMetrics indicates an expected call of ListAllMetrics.
func (mr *MockIMetricsStoreMockRecorder) ListAllMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAllMetrics", reflect.TypeOf((*MockIMetricsStore)(nil).ListAllMetrics), arg0)
}

// SaveMetric mocks base method.
func (m *MockIMetricsStore) SaveMetric(arg0 context.Context, arg1 models.Metric) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveMetric", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveMetric indicates an expected call of SaveMetric.
func (mr *MockIMetricsStoreMockRecorder) SaveMetric(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveMetric", reflect.TypeOf((*MockIMetricsStore)(nil).SaveMetric), arg0, arg1)
}

// SaveMetrics mocks base method.
func (m *MockIMetricsStore) SaveMetrics(arg0 context.Context, arg1 []models.Metric) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveMetrics", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveMetrics indicates an expected call of SaveMetrics.
func (mr *MockIMetricsStoreMockRecorder) SaveMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveMetrics", reflect.TypeOf((*MockIMetricsStore)(nil).SaveMetrics), arg0, arg1)
}
