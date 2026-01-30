package models

import "google.golang.org/protobuf/types/known/timestamppb"

// Product represents a product in the cashier system
type Product struct {
	ID         int                    `json:"id"`
	Name       string                 `json:"name"`
	Price      int                    `json:"price"`
	Stock      int                    `json:"stock"`
	CategoryID int                    `json:"category_id"`
	Category   *Category              `json:"category,omitempty"`
	DeletedAt  *timestamppb.Timestamp `json:"deleted_at"`
}
