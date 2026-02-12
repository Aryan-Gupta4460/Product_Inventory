package domain

import (
	"errors"
	"strings"
)

type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category"`
}

func (p Product) Validate() error {

	if strings.HasPrefix(strings.TrimSpace(p.Name), "--") {
		return errors.New("invalid product name")
	}
	if p.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if p.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	return nil
}
