package service

import "testing"

func TestNewUserService(t *testing.T) {
	svc := NewUserService()
	if svc == nil {
		t.Error("expected non-nil service")
	}
}
