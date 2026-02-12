package domain_test

import (
	"product_inventory/internal/domain"
	"testing"
)

func TestProductValidate_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		p       domain.Product
		wantErr bool
	}{
		{"valid", domain.Product{ID: "1", Name: "Good", Price: 1.0, Quantity: 1}, false},
		{"negative price", domain.Product{ID: "2", Name: "Bad", Price: -1, Quantity: 1}, true},
		{"negative quantity", domain.Product{ID: "3", Name: "Bad", Price: 1, Quantity: -5}, true},
		{"invalid name prefix", domain.Product{ID: "4", Name: "--bad", Price: 1, Quantity: 1}, true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.p.Validate(); (err != nil) != tc.wantErr {
				t.Fatalf("name=%s wantErr=%v got=%v", tc.name, tc.wantErr, err)
			}
		})
	}
}
