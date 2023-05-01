package storage

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Destinyxus/storeAPI/internal/models"
)

type Storage struct {
	db *gorm.DB
}

func NewStore() *Storage {
	return &Storage{}
}

type Result struct {
	CartID     uint
	CustomerID uint
}

type ProductCart struct {
	CartID    uint
	ProductID uint
}

func (s *Storage) Open() error {
	dsn := os.Getenv("DB")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	s.db = db

	err = s.db.AutoMigrate(&models.Customer{}, &models.Product{}, &models.ProductPhoto{}, &models.Cart{})

	return nil

}

func (s *Storage) CreateCart(customerID uint) error {
	// Check if customer already has a cart
	var existingCart models.Cart
	if err := s.db.Where("customer_id = ?", customerID).First(&existingCart).Error; err == nil {
		return fmt.Errorf("customer already has a cart")
	}

	// Create new cart for customer
	newCart := &models.Cart{CustomerID: customerID}
	result := s.db.Create(newCart)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *Storage) FindCartByCustomer(customerID uint) (*Result, error) {
	result := &Result{}

	if err := s.db.Table("customers").
		Select("carts.id as cart_id, customers.id as customer_id").
		Joins("LEFT JOIN carts ON carts.customer_id = customers.id").
		Where("customers.id = ?", customerID).
		Scan(result).Error; err != nil {
		return nil, err
	}
	return result, nil

}

func (s *Storage) AddProductToCart(cartID, productID uint) error {
	if err := s.db.Create(&ProductCart{
		CartID:    cartID,
		ProductID: productID,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (s *Storage) RetrieveProducts() (*models.Product, error) {
	productsList := new(models.Product)
	result := s.db.Find(productsList)

	return productsList, result.Error
}

func (s *Storage) CreateCustomer(customer *models.Customer) error {
	result := s.db.Create(&customer)
	if result.Error != nil {
		return result.Error
	}
	return nil

}

func (s *Storage) FindCustomerByEmail(email string) (*models.Customer, error) {

	user := new(models.Customer)
	if err := s.db.Where("email = ?", email).First(user).Error; err != nil {
		return nil, err
	}

	return user, nil

}
