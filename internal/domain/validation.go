package domain

import (
	"strings"
	"time"

	"guest-accounting/internal/rop"
)

// ValidateEmail returns an error if the email is not valid.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return NewDomainError("VALIDATION_FAILED", "invalid email")
	}
	return nil
}

// ValidatePassword returns an error if the password is too short.
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return NewDomainError("VALIDATION_FAILED", "password too short")
	}
	return nil
}

const guestDateLayout = "2006-01-02"

// ValidateGuest validates the guest record and returns it wrapped in Either.
// Left holds a VALIDATION_FAILED DomainError with field details on the first
// violated rule; Right holds the validated guest.
func ValidateGuest(g Guest) rop.Either[DomainError, Guest] {
	return rop.FlatMap(validateFullName(g), func(g Guest) rop.Either[DomainError, Guest] {
		return rop.FlatMap(validatePhone(g), func(g Guest) rop.Either[DomainError, Guest] {
			return rop.FlatMap(validateRoomNumber(g), func(g Guest) rop.Either[DomainError, Guest] {
				return rop.FlatMap(validateDates(g), func(g Guest) rop.Either[DomainError, Guest] {
					return validateNote(g)
				})
			})
		})
	})
}

func validateFullName(g Guest) rop.Either[DomainError, Guest] {
	if strings.TrimSpace(g.FullName) == "" {
		return rop.Left[DomainError, Guest](NewValidationError("full_name is required",
			FieldError{Field: "full_name", Message: "не может быть пустым"}))
	}
	return rop.Right[DomainError, Guest](g)
}

func validatePhone(g Guest) rop.Either[DomainError, Guest] {
	if strings.TrimSpace(g.Phone) == "" {
		return rop.Left[DomainError, Guest](NewValidationError("phone is required",
			FieldError{Field: "phone", Message: "не может быть пустым"}))
	}
	return rop.Right[DomainError, Guest](g)
}

func validateRoomNumber(g Guest) rop.Either[DomainError, Guest] {
	if strings.TrimSpace(g.RoomNumber) == "" {
		return rop.Left[DomainError, Guest](NewValidationError("room_number is required",
			FieldError{Field: "room_number", Message: "не может быть пустым"}))
	}
	return rop.Right[DomainError, Guest](g)
}

func validateDates(g Guest) rop.Either[DomainError, Guest] {
	checkIn, err := time.Parse(guestDateLayout, strings.TrimSpace(g.CheckIn))
	if err != nil {
		return rop.Left[DomainError, Guest](NewValidationError("check_in is invalid",
			FieldError{Field: "check_in", Message: "неверная дата заезда"}))
	}
	checkOut, err := time.Parse(guestDateLayout, strings.TrimSpace(g.CheckOut))
	if err != nil {
		return rop.Left[DomainError, Guest](NewValidationError("check_out is invalid",
			FieldError{Field: "check_out", Message: "неверная дата выезда"}))
	}
	if !checkOut.After(checkIn) {
		return rop.Left[DomainError, Guest](NewValidationError("check_out must be after check_in",
			FieldError{Field: "check_out", Message: "должна быть позже даты заезда"}))
	}
	return rop.Right[DomainError, Guest](g)
}

func validateNote(g Guest) rop.Either[DomainError, Guest] {
	if len(strings.TrimSpace(g.Note)) > 2000 {
		return rop.Left[DomainError, Guest](NewValidationError("note is too long",
			FieldError{Field: "note", Message: "не более 2000 символов"}))
	}
	return rop.Right[DomainError, Guest](g)
}
