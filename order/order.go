package order

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/sherwin-77/restaurant-golang-console/menu"
)

// Define base struct for any order item
type OrderItem interface {
	GetName() string
	GetPrice() float64
	GetQuantity() int
	GetTotal() float64
	GetMetadata() string
}

// Order struct
type MenuOrder struct {
	sync.Mutex
	Number         string
	items          []OrderItem
	OrderSignature string
}

func (m *MenuOrder) GetTotal() float64 {
	var total float64
	for _, item := range m.items {
		total += item.GetTotal()
	}
	return total
}

// Regenerate base64 signature for order
func (m *MenuOrder) recalculateSignature() {
	var metadata []string
	for _, item := range m.items {
		metadata = append(metadata, item.GetMetadata())
	}
	metadataStr := strings.Join(metadata, " | ")
	m.OrderSignature = base64.StdEncoding.EncodeToString([]byte(metadataStr))
}

func (m *MenuOrder) AddOrderItem(item OrderItem) {
	m.Lock()
	m.items = append(m.items, item)
	m.recalculateSignature()
	m.Unlock()
}

// Food order item
type FoodItem struct {
	menu.MenuItem
	Quantity int
}

func (f FoodItem) GetName() string {
	return f.MenuItem.Name
}

func (f FoodItem) GetPrice() float64 {
	return f.MenuItem.Price
}

func (f FoodItem) GetQuantity() int {
	return f.Quantity
}

func (f FoodItem) GetTotal() float64 {
	return f.MenuItem.Price * float64(f.GetQuantity())
}

func (f FoodItem) GetMetadata() string {
	return fmt.Sprintf("Name: %s, Price: %.2f, Quantity: %d", f.GetName(), f.GetPrice(), f.GetQuantity())
}

// Drink order item
type DrinkItem struct {
	menu.MenuItem
	Quantity int
	Refills  int
}

func (d DrinkItem) GetName() string {
	return d.MenuItem.Name
}

func (d DrinkItem) GetPrice() float64 {
	return d.MenuItem.Price
}

func (d DrinkItem) GetQuantity() int {
	return d.Quantity
}

func (d DrinkItem) GetTotal() float64 {
	return d.MenuItem.Price * float64(d.GetQuantity())
}

func (d DrinkItem) GetMetadata() string {
	return fmt.Sprintf("Name: %s, Price: %.2f, Quantity: %d, Refills: %d", d.GetName(), d.GetPrice(), d.GetQuantity(), d.Refills)
}
