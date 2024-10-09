package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type MenuItem struct {
	Name  string
	Price float64
}

type Order struct {
	Item  MenuItem
	Qty   int
	Total float64
}

type Menu struct {
	Items []MenuItem
}

var wg sync.WaitGroup

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi error:", r)
		}
		fmt.Println("Program selesai.")
	}()

	menu := InitMenu()
	var orders []Order
	orderChannel := make(chan Order, 5)

	go processOrders(orderChannel, &wg)

	for {
		fmt.Println("=== Sistem Manajemen Pesanan Restoran ===")
		fmt.Println("1. Tambah Pesanan")
		fmt.Println("2. Lihat Menu")
		fmt.Println("3. Keluar")
		fmt.Print("Pilih opsi: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			for {
				order := createOrder(menu)
				if order != nil {
					orders = append(orders, *order)
					orderChannel <- *order
				}
				var next string
				fmt.Print("Masukkan nama item (ketik 'selesai' untuk menyelesaikan): ")
				fmt.Scanln(&next)
				if next == "selesai" {
					close(orderChannel)
					break
				}
			}
			calculateTotalAndPayment(orders)
			wg.Wait() // Tunggu hingga semua goroutine selesai
			fmt.Println("Memproses pesanan")
		case 2:
			menu.PrintMenu()
		case 3:
			close(orderChannel)
			wg.Wait() // Tunggu hingga semua goroutine selesai
			fmt.Println("Selesai memproses semua pesanan.")
			os.Exit(0)
		default:
			fmt.Println("Opsi tidak valid!")
		}
	}
}

func InitMenu() *Menu {
	return &Menu{
		Items: []MenuItem{
			{"Nasi Goreng", 25000},
			{"Mie Goreng", 20000},
			{"Ayam Bakar", 30000},
		},
	}
}

func (m *Menu) AddItem(name string, price float64) {
	m.Items = append(m.Items, MenuItem{Name: name, Price: price})
}

func (m *Menu) PrintMenu() {
	fmt.Println("=== Menu Restoran ===")
	for i, item := range m.Items {
		fmt.Printf("%d. Nama: %s | Harga: %.2f\n", i+1, item.Name, item.Price)
	}
}

func createOrder(menu *Menu) *Order {
	var itemIndex, qty int
	menu.PrintMenu()
	fmt.Print("Pilih item berdasarkan nomor: ")
	fmt.Scanln(&itemIndex)

	if itemIndex <= 0 || itemIndex > len(menu.Items) {
		fmt.Println("Item tidak valid!")
		return nil
	}

	fmt.Print("Masukkan jumlah: ")
	fmt.Scanln(&qty)

	selectedItem := menu.Items[itemIndex-1]
	total := float64(qty) * selectedItem.Price

	order := Order{
		Item:  selectedItem,
		Qty:   qty,
		Total: total,
	}

	details := fmt.Sprintf("Pesanan: %s, Jumlah: %d, Total: %.2f", order.Item.Name, order.Qty, order.Total)
	encryptedDetails := encryptOrderDetails(details)
	fmt.Println("Pesanan (encoded base64):", encryptedDetails)

	return &order
}

func processOrders(orderChannel chan Order, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for order := range orderChannel {
		fmt.Printf("Memproses pesanan: %s | Jumlah: %d | Total: %.2f\n",
			order.Item.Name, order.Qty, order.Total)
		time.Sleep(2 * time.Second) // Simulasi proses pesanan
	}
}

func calculateTotalAndPayment(orders []Order) {
	var total float64
	fmt.Println("Pesanan Anda:")
	for _, order := range orders {
		fmt.Printf("- %s\n", order.Item.Name)
		total += order.Total
	}
	fmt.Printf("Total Harga: Rp%.2f\n", total)

	var payment float64
	for {
		fmt.Print("Masukkan jumlah yang dibayar: ")
		fmt.Scanln(&payment)

		if payment >= total {
			change := payment - total
			fmt.Printf("Jumlah yang dibayar valid. Kembalian: Rp%.2f\n", change)
			break
		} else {
			fmt.Println("Jumlah yang dibayar kurang, silakan coba lagi.")
		}
	}
}

func validatePrice(input string) (float64, error) {
	valid, _ := regexp.MatchString("^[0-9]+(\\.[0-9]{1,2})?$", input)
	if !valid {
		return 0, errors.New("harga tidak valid")
	}
	return strconv.ParseFloat(input, 64)
}

func encryptOrderDetails(details string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(details))
	return encoded
}
