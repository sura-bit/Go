package services

import (
	"errors"
	"sync"
	"time"

	"task4/models"
)


type LibraryManager interface {
	AddBook(book models.Book)
	RemoveBook(bookID int) error
	BorrowBook(bookID int, memberID int) error
	ReturnBook(bookID int, memberID int) error
	ListAvailableBooks() []models.Book
	ListBorrowedBooks(memberID int) []models.Book
	ReserveBook(bookID int, memberID int) error
	RegisterMember(m models.Member)
}

type reservation struct {
	MemberID int
	timer    *time.Timer
}

type ReservationRequest struct {
	BookID   int
	MemberID int
	RespCh   chan error
}

type Library struct {
	mu             sync.Mutex
	Books          map[int]models.Book
	Members        map[int]models.Member
	reservations   map[int]*reservation          
	reservationCh  chan ReservationRequest        
	cancelAllCh    chan struct{}                  
	reservationTTL time.Duration                  
}

func NewLibrary() *Library {
	l := &Library{
		Books:          make(map[int]models.Book),
		Members:        make(map[int]models.Member),
		reservations:   make(map[int]*reservation),
		reservationCh:  make(chan ReservationRequest, 128),
		cancelAllCh:    make(chan struct{}),
		reservationTTL: 5 * time.Second,
	}
	go l.startReservationWorker() 
	return l
}

func (l *Library) RegisterMember(m models.Member) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Members[m.ID] = m
}

func (l *Library) AddBook(book models.Book) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Books[book.ID] = book
}

func (l *Library) RemoveBook(bookID int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	book, ok := l.Books[bookID]
	if !ok {
		return errors.New("book not found")
	}
	if book.Status == "Borrowed" {
		return errors.New("cannot remove: book is borrowed")
	}
	if _, reserved := l.reservations[bookID]; reserved {
		return errors.New("cannot remove: book is reserved")
	}

	delete(l.Books, bookID)
	return nil
}

func (l *Library) BorrowBook(bookID int, memberID int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	book, ok := l.Books[bookID]
	if !ok {
		return errors.New("book not found")
	}
	member, ok := l.Members[memberID]
	if !ok {
		return errors.New("member not found")
	}


	if res, reserved := l.reservations[bookID]; reserved && res.MemberID != memberID {
		return errors.New("book is reserved by another member")
	}

	if book.Status == "Borrowed" {
		return errors.New("book already borrowed")
	}


	book.Status = "Borrowed"
	l.Books[bookID] = book
	member.BorrowedBooks = append(member.BorrowedBooks, book)
	l.Members[memberID] = member


	if res, reserved := l.reservations[bookID]; reserved && res.MemberID == memberID {
		if res.timer != nil {
			res.timer.Stop()
		}
		delete(l.reservations, bookID)
	}

	return nil
}

func (l *Library) ReturnBook(bookID int, memberID int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	member, ok := l.Members[memberID]
	if !ok {
		return errors.New("member not found")
	}

	book, ok := l.Books[bookID]
	if !ok {
		return errors.New("book not found")
	}

	found := false
	for i, b := range member.BorrowedBooks {
		if b.ID == bookID {
			member.BorrowedBooks = append(member.BorrowedBooks[:i], member.BorrowedBooks[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("book not borrowed by this member")
	}

	book.Status = "Available"
	l.Books[bookID] = book
	l.Members[memberID] = member
	return nil
}

func (l *Library) ListAvailableBooks() []models.Book {
	l.mu.Lock()
	defer l.mu.Unlock()

	var available []models.Book
	for id, b := range l.Books {
		if b.Status == "Available" {
			if _, reserved := l.reservations[id]; !reserved {
				available = append(available, b)
			}
		}
	}
	return available
}

func (l *Library) ListBorrowedBooks(memberID int) []models.Book {
	l.mu.Lock()
	defer l.mu.Unlock()

	member, ok := l.Members[memberID]
	if !ok {
		return nil
	}
	return append([]models.Book(nil), member.BorrowedBooks...)
}

func (l *Library) ReserveBook(bookID int, memberID int) error {
	resp := make(chan error, 1)
	req := ReservationRequest{BookID: bookID, MemberID: memberID, RespCh: resp}
	l.reservationCh <- req
	return <-resp
}
