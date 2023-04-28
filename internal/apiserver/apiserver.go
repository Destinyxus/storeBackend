package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"

	"github.com/Destinyxus/storeAPI/internal/authJWT"
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
	s.router.Middlewares()
	s.router.Get("/products", s.GetProducts())
	s.router.Post("/addProduct/{id}", s.authMiddleware(s.AddProductToCart()))
	s.router.Post("/createCustomer", s.CreateCustomer())
	s.router.Post("/createSession", s.AuthHandler())
	s.router.Post("/createCart", s.authMiddleware(s.CreateCart()))

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

		claims := request.Context().Value("claims").(*authJWT.MyCustomClaims)

		customerID, err := s.store.FindCustomerByEmail(claims.Email)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := s.store.CreateCart(customerID.ID); err != nil {
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
		claims := request.Context().Value("claims").(*authJWT.MyCustomClaims)

		customerID, err := s.store.FindCustomerByEmail(claims.Email)

		// Find the cart associated with the customer
		cart, err := s.store.FindCartByCustomer(customerID.ID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		id := chi.URLParam(request, "id")
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

		acc := models.NewCustomer(newCustomer.FirstName, newCustomer.LastName, newCustomer.Password, newCustomer.Phone, newCustomer.Email)

		err = s.store.CreateCustomer(acc)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		newCustomer.Sanitize()
		if err := writeToJson(writer, http.StatusCreated, newCustomer); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func (s *APIServer) AuthHandler() http.HandlerFunc {

	type auth1 struct {
		Email    string `json:"email"`
		Password string `json:"password,omitempty"`
	}

	logIn := new(auth1)

	return func(writer http.ResponseWriter, request *http.Request) {

		err := json.NewDecoder(request.Body).Decode(logIn)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		u, err := s.store.FindCustomerByEmail(logIn.Email)
		if err != nil || !u.CompareHash(logIn.Password) {

			if err := writeToJson(writer, http.StatusUnauthorized, "incorrect email or password"); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		tokenStr, err := authJWT.GenerateJWT(u.Email)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Authorization", "Bearer "+tokenStr)
	}
}

func (s *APIServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {
		authHeader := request.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			if err := writeToJson(writer, http.StatusUnauthorized, "incorrect token"); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := authJWT.ValidateToken(token)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(request.Context(), "claims", claims)

		next.ServeHTTP(writer, request.WithContext(ctx))
	}
}

func writeToJson(w http.ResponseWriter, code int, value interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(value)
}
