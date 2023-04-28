package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Destinyxus/storeAPI/internal/hashPass"
)

type CreateCustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password,omitempty"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

func (m *Customer) CompareHash(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(m.HashedPassword), []byte(password)) == nil
}

type CreateProductRequest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewCustomer(firstName, lastName, password, phone, email string) *Customer {
	cipherPassword, err := hashPass.CipherPassword(password)
	if err != nil {
		return nil
	}
	return &Customer{
		FirstName:      firstName,
		LastName:       lastName,
		HashedPassword: cipherPassword,
		Phone:          phone,
		Email:          email,
	}
}

func (m *CreateCustomerRequest) Sanitize() {
	m.Password = ""
}

type Customer struct {
	gorm.Model
	FirstName      string
	LastName       string
	HashedPassword string `gorm:"not null"`
	Phone          string
	Email          string `gorm:"unique;not null"`
	Cart           Cart
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
