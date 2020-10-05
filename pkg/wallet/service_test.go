package wallet

import (
	"reflect"
	"testing"
)

func TestService_RegisterAccount_success(t *testing.T) {
	svc := &Service{}
	expected, _ := svc.RegisterAccount("+992000000001")
	result, _ := svc.FindAccountByID(1)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected, result)
	}
}

func TestService_RegisterAccount_fail(t *testing.T) {
	svc := &Service{}
	svc.RegisterAccount("+992000000001")
	expected := ErrAccountNotFound
	_, result := svc.FindAccountByID(2)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected, result)
	}
}
