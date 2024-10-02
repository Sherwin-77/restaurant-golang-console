package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/sherwin-77/restaurant-golang-console/menu"
	"github.com/sherwin-77/restaurant-golang-console/order"
)

// Read integer from console
func readInt(reader *bufio.Reader) (int, error) {
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	input = strings.TrimSpace(input)
	number, err := strconv.Atoi(input)
	if err != nil {
		return 0, err
	}
	return number, nil
}

// Read float from console
func readFloat(reader *bufio.Reader) (float64, error) {
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	input = strings.TrimSpace(input)
	number, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, err
	}
	return number, nil
}

// Catch error and recover from panic
func catchError() {
	if r := recover(); r != nil {
		fmt.Println("Recovered from panic:", r)
	}
}

// Generate order based on item type
func generateOrder(item interface{}, orderChan chan<- order.OrderItem, wg *sync.WaitGroup) {
	defer wg.Done()
	switch item := item.(type) {
	case order.FoodItem:
		orderChan <- item
	case order.DrinkItem:
		orderChan <- item
	default:
		panic("Invalid item type")
	}
}

// Receive order from channel
func receiveOrder(orderChan <-chan order.OrderItem, menuOrder *order.MenuOrder) {
	for {
		item, ok := <-orderChan
		if !ok {
			break
		}
		fmt.Printf("\n[RESTAURANT] Received order (%s)\n", item.GetMetadata())
		menuOrder.AddOrderItem(item)
	}
}

func main() {
	runtime.GOMAXPROCS(2)
	defer catchError()
	defer fmt.Println("Thanks for visiting our restaurant. Have a great day!")

	orderList := map[string]menu.MenuItem{
		"burger":    {Name: "Burger", Price: 5.99, Type: menu.Food},
		"fries":     {Name: "Fries", Price: 3.99, Type: menu.Food},
		"pizza":     {Name: "Pizza", Price: 7.99, Type: menu.Food},
		"iced tea":  {Name: "Iced Tea", Price: 1.99, Type: menu.Drink},
		"coca cola": {Name: "Coca Cola", Price: 2.99, Type: menu.Drink},
		"pepsi":     {Name: "Pepsi", Price: 2.99, Type: menu.Drink},
	}

	cancelChan := make(chan os.Signal, 1)
	orderChan := make(chan order.OrderItem)
	signal.Notify(cancelChan, syscall.SIGINT, syscall.SIGTERM)

	menuOrder := order.MenuOrder{Number: "1"}
	reader := bufio.NewReader(os.Stdin)
	var wg sync.WaitGroup

	go receiveOrder(orderChan, &menuOrder)
	func() {
		for {
			fmt.Println("Menu List")
			for _, item := range orderList {
				fmt.Printf("%s: $%.2f\n", item.Name, item.Price)
			}
			fmt.Print("Enter your choice (type 'done' to complete order): ")

			choice, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Println("Please enter a valid choice.")
			}

			choice = strings.TrimSpace(choice)
			if choice == "done" {
				break
			}

			if item, ok := orderList[choice]; ok {
				fmt.Print("Enter quantity: ")
				quantity, err := readInt(reader)
				if err != nil || quantity < 1 {
					if err == io.EOF {
						return
					}
					fmt.Println("Please enter a valid quantity.")
					continue
				}

				switch item.Type {
				case menu.Food:
					wg.Add(1)
					go generateOrder(order.FoodItem{MenuItem: item, Quantity: quantity}, orderChan, &wg)
				case menu.Drink:
					fmt.Print("Enter refills: ")
					refills, err := readInt(reader)
					if err != nil || refills < 0 {
						if err == io.EOF {
							return
						}
						fmt.Println("Please enter a valid quantity.")
						continue
					}
					wg.Add(1)
					go generateOrder(order.DrinkItem{MenuItem: item, Quantity: quantity, Refills: refills}, orderChan, &wg)
				default:
					panic("Invalid item type")
				}
			} else {
				fmt.Println("Invalid choice. Please try again.")
			}
		}

		fmt.Printf("\nOrder Completed. Signature: %s\nTotal: $%.2f\n", menuOrder.OrderSignature, menuOrder.GetTotal())

		for {
			fmt.Print("Enter your amount: ")
			amount, err := readFloat(reader)
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Println("Please enter a valid amount.")
				continue
			}

			if amount >= menuOrder.GetTotal() {
				fmt.Printf("Change: $%.2f\n", amount-menuOrder.GetTotal())
				break
			} else {
				fmt.Println("Insufficient amount. Please try again.")
			}
		}
	}()

	fmt.Println("\nWaiting for order to be processed...")
	wg.Wait()
	close(orderChan)

	select {
	case signl, ok := <-cancelChan:
		if ok {
			fmt.Printf("Received signal: %s\n", signl)
		}
	default:
		fmt.Println("Program completed successfully.")
	}
	close(cancelChan)
}
