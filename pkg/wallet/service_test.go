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

func TestService_Reject_success(t *testing.T) {
	svc := &Service{}
	svc.RegisterAccount("+992000000001")
	svc.Deposit(1, 20)
	expected_payment, expected_error := svc.Pay(1, 10, "food")
	result_payment, result_error := svc.FindPaymentByID(expected_payment.ID)

	if !reflect.DeepEqual(expected_payment, result_payment) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected_payment, result_payment)
	}

	if !reflect.DeepEqual(expected_error, result_error) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected_error, result_error)
	}
}

func TestService_Reject_fail(t *testing.T) {
	svc := &Service{}
	svc.RegisterAccount("+992000000001")
	svc.Deposit(1, 20)
	expected_error := ErrPaymentNotFound
	_, result_error := svc.FindPaymentByID("second")

	if !reflect.DeepEqual(expected_error, result_error) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected_error, result_error)
	}
}