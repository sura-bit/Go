package controllers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"task4/models"
	"task4/services"
)

func readLine(r *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	text, _ := r.ReadString('\n')
	return strings.TrimSpace(text)
}

func asInt(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

func simulateConcurrentReservations(lib *services.Library) {
	fmt.Println("\n-- Simulating concurrent reservations for BookID=1 by Members 1..5 --")
	var wg sync.WaitGroup
	for mid := 1; mid <= 5; mid++ {
		wg.Add(1)
		go func(memberID int) {
			defer wg.Done()
			err := lib.ReserveBook(1, memberID)
			if err != nil {
				fmt.Printf("[Member %d] Reserve error: %v\n", memberID, err)
			} else {
				fmt.Printf("[Member %d] Reserved successfully.\n", memberID)
			}
		}(mid)
	}
	wg.Wait()
	fmt.Println("-- Done. Try borrowing as the winning member before 5s passes! --")
}

func RunLibrarySystem() {
	r := bufio.NewReader(os.Stdin)
	library := services.NewLibrary()

	// Seed data
	library.RegisterMember(models.Member{ID: 1, Name: "Alice"})
	library.RegisterMember(models.Member{ID: 2, Name: "Bob"})
	library.RegisterMember(models.Member{ID: 3, Name: "Charlie"})
	library.RegisterMember(models.Member{ID: 4, Name: "Dana"})
	library.RegisterMember(models.Member{ID: 5, Name: "Evan"})

	library.AddBook(models.Book{ID: 1, Title: "The Go Programming Language", Author: "Donovan & Kernighan", Status: "Available"})
	library.AddBook(models.Book{ID: 2, Title: "Concurrency in Go", Author: "Katherine Cox-Buday", Status: "Available"})

	for {
		fmt.Println("\n=== Library Management System (Concurrent) ===")
		fmt.Println("1. Add Book")
		fmt.Println("2. Remove Book")
		fmt.Println("3. Borrow Book")
		fmt.Println("4. Return Book")
		fmt.Println("5. List Available Books")
		fmt.Println("6. List Borrowed Books (by Member)")
		fmt.Println("7. Reserve Book")
		fmt.Println("8. Simulate Concurrent Reservations")
		fmt.Println("9. Exit")

		choice := asInt(readLine(r, "Enter choice: "))

		switch choice {
		case 1:
			id := asInt(readLine(r, "Book ID: "))
			title := readLine(r, "Title: ")
			author := readLine(r, "Author: ")
			library.AddBook(models.Book{ID: id, Title: title, Author: author, Status: "Available"})
			fmt.Println("Book added.")
		case 2:
			id := asInt(readLine(r, "Book ID to remove: "))
			if err := library.RemoveBook(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Book removed.")
			}
		case 3:
			bid := asInt(readLine(r, "Book ID to borrow: "))
			mid := asInt(readLine(r, "Member ID: "))
			if err := library.BorrowBook(bid, mid); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Borrowed successfully.")
			}
		case 4:
			bid := asInt(readLine(r, "Book ID to return: "))
			mid := asInt(readLine(r, "Member ID: "))
			if err := library.ReturnBook(bid, mid); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Returned successfully.")
			}
		case 5:
			fmt.Println("\nAvailable Books:")
			for _, b := range library.ListAvailableBooks() {
				fmt.Printf("ID: %d | %s — %s\n", b.ID, b.Title, b.Author)
			}
		case 6:
			mid := asInt(readLine(r, "Member ID: "))
			books := library.ListBorrowedBooks(mid)
			if len(books) == 0 {
				fmt.Println("No borrowed books.")
			} else {
				for _, b := range books {
					fmt.Printf("ID: %d | %s — %s\n", b.ID, b.Title, b.Author)
				}
			}
		case 7:
			bid := asInt(readLine(r, "Book ID to reserve: "))
			mid := asInt(readLine(r, "Member ID: "))
			if err := library.ReserveBook(bid, mid); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Reserved successfully (auto-cancels in 5s if not borrowed).")
			}
		case 8:
			simulateConcurrentReservations(library)
		case 9:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice.")
		}
	}
}
