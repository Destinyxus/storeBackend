package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"

	"github.com/Destinyxus/storeAPI/internal/models"
	"github.com/Destinyxus/storeAPI/internal/storage"
)

type APIServer struct {
	router *chi.Mux
	store  *storage.Storage
}

func NewAPIServer() *APIServer {

	return &APIServer{
		router: chi.NewRouter(),
		store:  storage.NewStore(),
	}
}

func (s *APIServer) Run() error {
	s.configureStore()
	s.configureRouter()

	http.ListenAndServe("localhost:8080", s.router)
	return nil
}

func (s *APIServer) configureRouter() error {
	s.router.Get("/products", s.GetProducts())
	s.router.Post("/addProduct/{productID}", s.AddProductToCart())
	s.router.Post("/createCustomer", s.CreateCustomer())
	s.router.Post("/createCart", s.CreateCart())

	return nil
}

func (s *APIServer) configureStore() error {
	s.store.Open()
	return nil
}
func (s *APIServer) GetProducts() http.HandlerFunc {
	prods, err := s.store.RetrieveProducts()
	if err != nil {
		fmt.Println(err)
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(prods.Name))

	}
}

func (s *APIServer) CreateCart() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		sessionID, err := request.Cookie("session_id")
		if err != nil {
			http.Error(writer, "session ID cookie not found", http.StatusBadRequest)
			return
		}

		customerID, err := s.store.FindCustomerBySession(sessionID.Value)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := s.store.CreateCart(customerID); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return a success response to the client
		writer.WriteHeader(http.StatusCreated)
		writer.Write([]byte("Cart created successfully"))
	}
}

func (s *APIServer) AddProductToCart() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		sessionID, err := request.Cookie("session_id")
		if err != nil {
			http.Error(writer, "session ID cookie not found", http.StatusBadRequest)
			return
		}

		// Find the customer associated with the session ID
		customerID, err := s.store.FindCustomerBySession(sessionID.Value)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Find the cart associated with the customer
		cart, err := s.store.FindCartByCustomer(customerID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(cart.CartID)

		id := chi.URLParam(request, "productID")
		idd, err := strconv.Atoi(id)
		err = s.store.AddProductToCart(cart.CartID, uint(idd))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return a success response to the client
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("Product added to cart successfully"))

	}
}

func (s *APIServer) CreateCustomer() http.HandlerFunc {
	newCustomer := new(models.CreateCustomerRequest)

	return func(writer http.ResponseWriter, request *http.Request) {
		err := json.NewDecoder(request.Body).Decode(newCustomer)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		sessionID := uuid.New()

		acc := models.NewCustomer(newCustomer.FirstName, newCustomer.LastName, newCustomer.Phone, newCustomer.Email)

		customerID, err := s.store.CreateCustomer(acc)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		session := &models.Session{
			SessionID:  sessionID.String(),
			CustomerID: customerID,
		}

		if err := s.store.CreateSession(session); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		expiration := time.Now().Add(24 * time.Hour)
		cookie := &http.Cookie{
			Name:    "session_id",
			Value:   session.SessionID,
			Expires: expiration,
		}

		http.SetCookie(writer, cookie)

		if err := writeToJson(writer, http.StatusCreated, acc); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func writeToJson(w http.ResponseWriter, code int, value interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(value)
}
