package service

import (
	"guest-accounting/internal/domain"
)

// GuestService handles guest CRUD operations.
type GuestService interface {
	Create(userID int64, g domain.Guest) (domain.Guest, error)
	Get(id, userID int64) (domain.Guest, error)
	List(userID int64, page, limit int) (domain.GuestList, error)
	Search(userID, query string, page, limit int) (domain.GuestList, error)
	Update(id, userID int64, g domain.Guest) (domain.Guest, error)
	Delete(id, userID int64) error
}

type guestService struct{}

// NewGuestService creates a stub guest service.
func NewGuestService() GuestService {
	return &guestService{}
}

func (s *guestService) Create(userID int64, g domain.Guest) (domain.Guest, error) {
	return domain.Guest{}, domain.ErrNotImplemented
}

func (s *guestService) Get(id, userID int64) (domain.Guest, error) {
	return domain.Guest{}, domain.ErrNotImplemented
}

func (s *guestService) List(userID int64, page, limit int) (domain.GuestList, error) {
	return domain.GuestList{}, nil
}

func (s *guestService) Search(userID, query string, page, limit int) (domain.GuestList, error) {
	return domain.GuestList{}, nil
}

func (s *guestService) Update(id, userID int64, g domain.Guest) (domain.Guest, error) {
	return domain.Guest{}, domain.ErrNotImplemented
}

func (s *guestService) Delete(id, userID int64) error {
	return domain.ErrNotImplemented
}
