package apperrors

import (
	"errors"
	"fmt"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrDuplicateProduct = errors.New("duplicate product")
	ErrInvalidProduct   = errors.New("invalid product")
)

type ProductNotFoundError struct {
	ID string
}

func (e *ProductNotFoundError) Error() string {
	return fmt.Sprintf("product with ID %s not found", e.ID)
}

func (e *ProductNotFoundError) Unwrap() error {
	return ErrProductNotFound
}

type DuplicateProductError struct {
	ID string
}

func (e *DuplicateProductError) Error() string {
	return fmt.Sprintf("product with ID %s already exists", e.ID)
}

func (e *DuplicateProductError) Unwrap() error {
	return ErrDuplicateProduct
}

type InvalidProductError struct {
	Reason string
}

func (e *InvalidProductError) Error() string {
	return fmt.Sprintf("invalid product: %s", e.Reason)
}

func (e *InvalidProductError) Unwrap() error {
	return ErrInvalidProduct
}
