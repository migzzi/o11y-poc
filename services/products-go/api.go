package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var pricingServiceURL = os.Getenv("SALES_SERVICE_URL")
var pricingServiceClient = &http.Client{
	Transport: otelhttp.NewTransport(http.DefaultTransport),
}

type AddProductDto struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type ProductPricingInfo struct {
	ID       string  `json:"productId"`
	Price    float64 `json:"price"`
	Discount float64 `json:"discount"`
	Total    float64 `json:"total"`
}

type ProductDetailsDto struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Pricing     ProductPricingInfo `json:"pricing"`
}

type Product struct {
	ID          int
	Name        string
	Description string
}

type GenericError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

var products = []*Product{
	{ID: 1, Name: "Product 1", Description: "Description 1"},
	{ID: 2, Name: "Product 2", Description: "Description 2"},
	{ID: 3, Name: "Product 3", Description: "Description 3"},
}

func getProductsPrices(ctx context.Context, ids []string) (map[string]*ProductPricingInfo, error) {
	// Call the pricing service to get the prices of the products
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/pricing/%s", pricingServiceURL, strings.Join(ids, ",")), nil)
	if err != nil {
		return nil, err
	}

	log.Printf("Fetching prices for products %s", strings.Join(ids, ","))

	req = req.WithContext(ctx)
	resp, err := pricingServiceClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pricings []*ProductPricingInfo
	err = json.NewDecoder(resp.Body).Decode(&pricings)
	if err != nil {
		return nil, err
	}

	var prices = make(map[string]*ProductPricingInfo)
	for _, p := range pricings {
		prices[p.ID] = p
	}

	log.Printf("Prices fetched %v", prices)
	return prices, nil
}

func getProductWithFullDetails(ctx context.Context, ids []string) ([]*ProductDetailsDto, error) {
	prices, err := getProductsPrices(ctx, ids)
	if err != nil {
		return nil, err
	}

	productsDetails := make([]*ProductDetailsDto, len(products))
	for i, p := range products {
		pricing, ok := prices[strconv.Itoa(p.ID)]
		if !ok {
			return nil, fmt.Errorf("price not found for product %d", p.ID)
		}
		productsDetails[i] = &ProductDetailsDto{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Pricing:     *pricing,
		}
	}
	return productsDetails, nil
}

func sendError(w http.ResponseWriter, err error, status int) {
	errMsg := GenericError{Message: err.Error(), Status: status}
	errJSON, _ := json.Marshal(errMsg)
	http.Error(w, string(errJSON), status)
}

// APIs handlers.
func productsHandler(w http.ResponseWriter, r *http.Request) {

	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	ids := make([]string, len(products))
	for i, p := range products {
		ids[i] = strconv.Itoa(p.ID)
	}

	prices, err := getProductWithFullDetails(r.Context(), ids)
	if err != nil {
		sendError(w, err, http.StatusInternalServerError)
		return
	}
	// Marshal the products into a JSON string
	productsJSON, err := json.Marshal(prices)
	if err != nil {
		sendError(w, err, http.StatusInternalServerError)
		return
	}

	// Write the JSON string to the response
	w.Write(productsJSON)
}

func getProductByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Get the product ID from the URL
	id := r.PathValue("id")

	// Find the product with the given ID
	var product *Product = nil
	for _, p := range products {
		if id == strconv.Itoa(p.ID) {
			product = p
			break
		}
	}
	if product == nil {
		sendError(w, fmt.Errorf("product not found"), http.StatusNotFound)
		return
	}
	// Fetch the product pricing details
	productDetails, err := getProductWithFullDetails(r.Context(), []string{id})
	if err != nil {
		sendError(w, err, http.StatusInternalServerError)
		return
	}

	// Marshal the product into a JSON string
	productJSON, err := json.Marshal(productDetails[0])
	if err != nil {
		sendError(w, fmt.Errorf("Error fetching price. %w", err), http.StatusInternalServerError)
		return
	}

	// Write the JSON string to the response
	w.Write(productJSON)
}

func addProductHandler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create a new product from the request body
	var productDto AddProductDto
	err := json.NewDecoder(r.Body).Decode(&productDto)
	if err != nil {
		sendError(w, err, http.StatusBadRequest)
		return
	}
	product := Product{
		ID:          idGen.NextID(),
		Name:        productDto.Name,
		Description: productDto.Description,
	}

	// Append the new product to the products slice
	products = append(products, &product)

	// Marshal the new product into a JSON string
	productJSON, err := json.Marshal(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the JSON string to the response
	w.Write(productJSON)
}
