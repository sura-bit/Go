package controllers

import (
	"fmt"
	"task3/models"
	"task3/services"
)


func RunLibrarySystem() {
	library := services.NewLibrary()

	library.Members[1] = models.Member{ID: 1, Name: "Alice"}
	library.Members[2] = models.Member{ID: 2, Name: "Bob"}

	for {
		fmt.Println("\n=== Library Management System ===")
		fmt.Println("1. Add Book")
		fmt.Println("2. Remove Book")
		fmt.Println("3. Borrow Book")
		fmt.Println("4. Return Book")
		fmt.Println("5. List Available Books")
		fmt.Println("6. List Borrowed Books")
		fmt.Println("7. Exit")

		var choice int
		fmt.Print("Enter your choice: ")
		fmt.Scan(&choice)

		switch choice {
		case 1:
			var id int
			var title, author string
			fmt.Print("Enter Book ID: ")
			fmt.Scan(&id)
			fmt.Print("Enter Title: ")
			fmt.Scan(&title)
			fmt.Print("Enter Author: ")
			fmt.Scan(&author)

			library.AddBook(models.Book{ID: id, Title: title, Author: author, Status: "Available"})
			fmt.Println("Book added successfully.")

		case 2:
			var id int
			fmt.Print("Enter Book ID to remove: ")
			fmt.Scan(&id)
			library.RemoveBook(id)
			fmt.Println("Book removed.")

		case 3:
			var bookID, memberID int
			fmt.Print("Enter Book ID to borrow: ")
			fmt.Scan(&bookID)
			fmt.Print("Enter Member ID: ")
			fmt.Scan(&memberID)

			err := library.BorrowBook(bookID, memberID)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Book borrowed successfully.")
			}

		case 4:
			var bookID, memberID int
			fmt.Print("Enter Book ID to return: ")
			fmt.Scan(&bookID)
			fmt.Print("Enter Member ID: ")
			fmt.Scan(&memberID)

			err := library.ReturnBook(bookID, memberID)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Book returned successfully.")
			}

		case 5:
			fmt.Println("Available Books:")
			for _, b := range library.ListAvailableBooks() {
				fmt.Printf("ID: %d, Title: %s, Author: %s\n", b.ID, b.Title, b.Author)
			}

		case 6:
			var memberID int
			fmt.Print("Enter Member ID: ")
			fmt.Scan(&memberID)
			books := library.ListBorrowedBooks(memberID)
			if len(books) == 0 {
				fmt.Println("No borrowed books.")
			} else {
				for _, b := range books {
					fmt.Printf("ID: %d, Title: %s, Author: %s\n", b.ID, b.Title, b.Author)
				}
			}

		case 7:
			fmt.Println("Exiting system. Goodbye!")
			return
		default:
			fmt.Println("Invalid choice.")
		}
	}
}
