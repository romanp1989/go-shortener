// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/romanp1989/go-shortener/internal/models (interfaces: Storage)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	uuid "github.com/gofrs/uuid"
	gomock "github.com/golang/mock/gomock"
	models "github.com/romanp1989/go-shortener/internal/models"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// DeleteBatch mocks base method.
func (m *MockStorage) DeleteBatch(arg0 context.Context, arg1 *uuid.UUID, arg2 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBatch", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteBatch indicates an expected call of DeleteBatch.
func (mr *MockStorageMockRecorder) DeleteBatch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBatch", reflect.TypeOf((*MockStorage)(nil).DeleteBatch), arg0, arg1, arg2)
}

// Get mocks base method.
func (m *MockStorage) Get(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStorageMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStorage)(nil).Get), arg0)
}

// GetAllUrlsByUser mocks base method.
func (m *MockStorage) GetAllUrlsByUser(arg0 context.Context, arg1 *uuid.UUID) ([]models.StorageURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllUrlsByUser", arg0, arg1)
	ret0, _ := ret[0].([]models.StorageURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllUrlsByUser indicates an expected call of GetAllUrlsByUser.
func (mr *MockStorageMockRecorder) GetAllUrlsByUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllUrlsByUser", reflect.TypeOf((*MockStorage)(nil).GetAllUrlsByUser), arg0, arg1)
}

// GetStats mocks base method.
func (m *MockStorage) GetStats(arg0 context.Context) (models.StorageStats, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStats", arg0)
	ret0, _ := ret[0].(models.StorageStats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStats indicates an expected call of GetStats.
func (mr *MockStorageMockRecorder) GetStats(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStats", reflect.TypeOf((*MockStorage)(nil).GetStats), arg0)
}

// Ping mocks base method.
func (m *MockStorage) Ping(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockStorageMockRecorder) Ping(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockStorage)(nil).Ping), arg0)
}

// Save mocks base method.
func (m *MockStorage) Save(arg0 context.Context, arg1, arg2 string, arg3 *uuid.UUID) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Save indicates an expected call of Save.
func (mr *MockStorageMockRecorder) Save(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockStorage)(nil).Save), arg0, arg1, arg2, arg3)
}

// SaveBatch mocks base method.
func (m *MockStorage) SaveBatch(arg0 context.Context, arg1 []models.StorageURL, arg2 *uuid.UUID) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveBatch", arg0, arg1, arg2)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveBatch indicates an expected call of SaveBatch.
func (mr *MockStorageMockRecorder) SaveBatch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveBatch", reflect.TypeOf((*MockStorage)(nil).SaveBatch), arg0, arg1, arg2)
}
