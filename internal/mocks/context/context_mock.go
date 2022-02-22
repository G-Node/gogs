// Code generated by MockGen. DO NOT EDIT.
// Source: internal/context/context.go

// Package mock_context is a generated GoMock package.
package mock_context

import (
	reflect "reflect"

	session "github.com/go-macaron/session"
	gomock "github.com/golang/mock/gomock"
	context "github.com/ivis-yoshida/gogs/internal/context"
	db "github.com/ivis-yoshida/gogs/internal/db"
)

// MockAbstructContext is a mock of AbstructContext interface.
type MockAbstructContext struct {
	ctrl     *gomock.Controller
	recorder *MockAbstructContextMockRecorder
}

// MockAbstructContextMockRecorder is the mock recorder for MockAbstructContext.
type MockAbstructContextMockRecorder struct {
	mock *MockAbstructContext
}

// NewMockAbstructContext creates a new mock instance.
func NewMockAbstructContext(ctrl *gomock.Controller) *MockAbstructContext {
	mock := &MockAbstructContext{ctrl: ctrl}
	mock.recorder = &MockAbstructContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAbstructContext) EXPECT() *MockAbstructContextMockRecorder {
	return m.recorder
}

// CallData mocks base method.
func (m *MockAbstructContext) CallData() map[string]interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CallData")
	ret0, _ := ret[0].(map[string]interface{})
	return ret0
}

// CallData indicates an expected call of CallData.
func (mr *MockAbstructContextMockRecorder) CallData() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CallData", reflect.TypeOf((*MockAbstructContext)(nil).CallData))
}

// Error mocks base method.
func (m *MockAbstructContext) Error(err error, msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", err, msg)
}

// Error indicates an expected call of Error.
func (mr *MockAbstructContextMockRecorder) Error(err, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockAbstructContext)(nil).Error), err, msg)
}

// GetFlash mocks base method.
func (m *MockAbstructContext) GetFlash() *session.Flash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFlash")
	ret0, _ := ret[0].(*session.Flash)
	return ret0
}

// GetFlash indicates an expected call of GetFlash.
func (mr *MockAbstructContextMockRecorder) GetFlash() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFlash", reflect.TypeOf((*MockAbstructContext)(nil).GetFlash))
}

// GetRepo mocks base method.
func (m *MockAbstructContext) GetRepo() context.AbstructCtxRepository {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRepo")
	ret0, _ := ret[0].(context.AbstructCtxRepository)
	return ret0
}

// GetRepo indicates an expected call of GetRepo.
func (mr *MockAbstructContextMockRecorder) GetRepo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRepo", reflect.TypeOf((*MockAbstructContext)(nil).GetRepo))
}

// GetUser mocks base method.
func (m *MockAbstructContext) GetUser() *db.User {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser")
	ret0, _ := ret[0].(*db.User)
	return ret0
}

// GetUser indicates an expected call of GetUser.
func (mr *MockAbstructContextMockRecorder) GetUser() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockAbstructContext)(nil).GetUser))
}

// PageIs mocks base method.
func (m *MockAbstructContext) PageIs(name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PageIs", name)
}

// PageIs indicates an expected call of PageIs.
func (mr *MockAbstructContextMockRecorder) PageIs(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PageIs", reflect.TypeOf((*MockAbstructContext)(nil).PageIs), name)
}

// Query mocks base method.
func (m *MockAbstructContext) Query(name string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", name)
	ret0, _ := ret[0].(string)
	return ret0
}

// Query indicates an expected call of Query.
func (mr *MockAbstructContextMockRecorder) Query(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockAbstructContext)(nil).Query), name)
}

// QueryEscape mocks base method.
func (m *MockAbstructContext) QueryEscape(name string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryEscape", name)
	ret0, _ := ret[0].(string)
	return ret0
}

// QueryEscape indicates an expected call of QueryEscape.
func (mr *MockAbstructContextMockRecorder) QueryEscape(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryEscape", reflect.TypeOf((*MockAbstructContext)(nil).QueryEscape), name)
}

// QueryInt mocks base method.
func (m *MockAbstructContext) QueryInt(name string) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryInt", name)
	ret0, _ := ret[0].(int)
	return ret0
}

// QueryInt indicates an expected call of QueryInt.
func (mr *MockAbstructContextMockRecorder) QueryInt(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryInt", reflect.TypeOf((*MockAbstructContext)(nil).QueryInt), name)
}

// Redirect mocks base method.
func (m *MockAbstructContext) Redirect(location string, status ...int) {
	m.ctrl.T.Helper()
	varargs := []interface{}{location}
	for _, a := range status {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Redirect", varargs...)
}

// Redirect indicates an expected call of Redirect.
func (mr *MockAbstructContextMockRecorder) Redirect(location interface{}, status ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{location}, status...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Redirect", reflect.TypeOf((*MockAbstructContext)(nil).Redirect), varargs...)
}

// RequireHighlightJS mocks base method.
func (m *MockAbstructContext) RequireHighlightJS() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RequireHighlightJS")
}

// RequireHighlightJS indicates an expected call of RequireHighlightJS.
func (mr *MockAbstructContextMockRecorder) RequireHighlightJS() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequireHighlightJS", reflect.TypeOf((*MockAbstructContext)(nil).RequireHighlightJS))
}

// RequireSimpleMDE mocks base method.
func (m *MockAbstructContext) RequireSimpleMDE() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RequireSimpleMDE")
}

// RequireSimpleMDE indicates an expected call of RequireSimpleMDE.
func (mr *MockAbstructContextMockRecorder) RequireSimpleMDE() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequireSimpleMDE", reflect.TypeOf((*MockAbstructContext)(nil).RequireSimpleMDE))
}

// Success mocks base method.
func (m *MockAbstructContext) Success(name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Success", name)
}

// Success indicates an expected call of Success.
func (mr *MockAbstructContextMockRecorder) Success(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Success", reflect.TypeOf((*MockAbstructContext)(nil).Success), name)
}

// Tr mocks base method.
func (m *MockAbstructContext) Tr(arg0 string, arg1 ...interface{}) string {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Tr", varargs...)
	ret0, _ := ret[0].(string)
	return ret0
}

// Tr indicates an expected call of Tr.
func (mr *MockAbstructContextMockRecorder) Tr(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tr", reflect.TypeOf((*MockAbstructContext)(nil).Tr), varargs...)
}

// UserID mocks base method.
func (m *MockAbstructContext) UserID() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UserID")
	ret0, _ := ret[0].(int64)
	return ret0
}

// UserID indicates an expected call of UserID.
func (mr *MockAbstructContextMockRecorder) UserID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserID", reflect.TypeOf((*MockAbstructContext)(nil).UserID))
}
