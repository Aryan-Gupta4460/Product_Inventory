package apperrors_test

import (
	"errors"
	"testing"

	apperrors "product_inventory/internal/errors"
)

func TestCustomErrors_IsAndAs(t *testing.T) {
	e1 := &apperrors.ProductNotFoundError{ID: "x"}
	if !errors.Is(e1, apperrors.ErrProductNotFound) {
		t.Fatalf("expected Is to match ErrProductNotFound")
	}

	var pn *apperrors.ProductNotFoundError
	if !errors.As(e1, &pn) || pn.ID != "x" {
		t.Fatalf("errors.As failed to retrieve ProductNotFoundError")
	}

	e2 := &apperrors.InvalidProductError{Reason: "bad"}
	if !errors.Is(e2, apperrors.ErrInvalidProduct) {
		t.Fatalf("expected Is to match ErrInvalidProduct")
	}

	e3 := &apperrors.DuplicateProductError{ID: "d"}
	if !errors.Is(e3, apperrors.ErrDuplicateProduct) {
		t.Fatalf("expected Is to match ErrDuplicateProduct")
	}
}
