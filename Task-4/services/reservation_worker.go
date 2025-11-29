package services

import (
	"errors"
	"time"
)


func (l *Library) startReservationWorker() {
	for {
		select {
		case req := <-l.reservationCh:
			l.mu.Lock()

			book, exists := l.Books[req.BookID]
			if !exists {
				l.mu.Unlock()
				req.RespCh <- errors.New("book not found")
				continue
			}
			if _, mExists := l.Members[req.MemberID]; !mExists {
				l.mu.Unlock()
				req.RespCh <- errors.New("member not found")
				continue
			}
			if book.Status == "Borrowed" {
				l.mu.Unlock()
				req.RespCh <- errors.New("book already borrowed")
				continue
			}
			if _, reserved := l.reservations[req.BookID]; reserved {
				l.mu.Unlock()
				req.RespCh <- errors.New("book already reserved")
				continue
			}

			timer := time.AfterFunc(l.reservationTTL, func() {
				l.mu.Lock()
				defer l.mu.Unlock()
				if r, ok := l.reservations[req.BookID]; ok && r.MemberID == req.MemberID {
					if b, ok2 := l.Books[req.BookID]; ok2 && b.Status != "Borrowed" {
						delete(l.reservations, req.BookID)
					}
				}
			})

			l.reservations[req.BookID] = &reservation{
				MemberID: req.MemberID,
				timer:    timer,
			}
			l.mu.Unlock()

			req.RespCh <- nil

		case <-l.cancelAllCh:
			return
		}
	}
}
