// Code generated by MockGen. DO NOT EDIT.
// Source: folder.go
//
// Generated by this command:
//
//	mockgen -source=folder.go -destination=folder_mock.go -package=repository
//
// Package repository is a generated GoMock package.
package repository

import (
	reflect "reflect"
	entity "resource/domain/entity"

	gomock "go.uber.org/mock/gomock"
)

// MockFolderRepo is a mock of FolderRepo interface.
type MockFolderRepo struct {
	ctrl     *gomock.Controller
	recorder *MockFolderRepoMockRecorder
}

// MockFolderRepoMockRecorder is the mock recorder for MockFolderRepo.
type MockFolderRepoMockRecorder struct {
	mock *MockFolderRepo
}

// NewMockFolderRepo creates a new mock instance.
func NewMockFolderRepo(ctrl *gomock.Controller) *MockFolderRepo {
	mock := &MockFolderRepo{ctrl: ctrl}
	mock.recorder = &MockFolderRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFolderRepo) EXPECT() *MockFolderRepoMockRecorder {
	return m.recorder
}

// CreateFolder mocks base method.
func (m *MockFolderRepo) CreateFolder(folder entity.Folder) (*entity.Folder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFolder", folder)
	ret0, _ := ret[0].(*entity.Folder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateFolder indicates an expected call of CreateFolder.
func (mr *MockFolderRepoMockRecorder) CreateFolder(folder any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFolder", reflect.TypeOf((*MockFolderRepo)(nil).CreateFolder), folder)
}

// GetSubFolders mocks base method.
func (m *MockFolderRepo) GetSubFolders(folderID uint) ([]*entity.Folder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubFolders", folderID)
	ret0, _ := ret[0].([]*entity.Folder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubFolders indicates an expected call of GetSubFolders.
func (mr *MockFolderRepoMockRecorder) GetSubFolders(folderID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubFolders", reflect.TypeOf((*MockFolderRepo)(nil).GetSubFolders), folderID)
}
