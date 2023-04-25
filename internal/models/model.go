package models

import (
	"gorm.io/gorm"
)

type CreateCustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type CreateProductRequest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewCustomer(firstName, lastName, phone, email string) *Customer {
	return &Customer{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Email:     email,
	}
}

type Customer struct {
	gorm.Model
	FirstName string
	LastName  string
	Phone     string
	Email     string `gorm:"unique;not null"`
	Cart      Cart
	Sessions  []Session
}

type Cart struct {
	gorm.Model
	CustomerID uint
	Products   []Product `gorm:"many2many:product_carts;"`
}

type Product struct {
	gorm.Model
	Name         string
	Description  string
	Price        float64
	ProductPhoto []ProductPhoto
	Carts        []Cart `gorm:"many2many:product_carts;"`
}

type ProductPhoto struct {
	gorm.Model
	Photo     string
	ProductID uint
}

type Session struct {
	gorm.Model
	CustomerID uint
	SessionID  string
}
