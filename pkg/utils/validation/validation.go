package validation

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type Errors []error

func (e *Errors) Add(err error) {
	if err != nil {
		*e = append(*e, err)
	}
}

type ValidationError struct {
	Errs []error
}

func (v ValidationError) Error() string {
	if len(v.Errs) == 0 {
		return ""
	}
	return errors.Join(v.Errs...).Error()
}

func (v ValidationError) StatusCode() int {
	return http.StatusBadRequest
}

func (v ValidationError) Details() any {
	msgs := make([]string, len(v.Errs))
	for i, e := range v.Errs {
		msgs[i] = e.Error()
	}
	return msgs
}

func (e Errors) Result() error {
	if len(e) == 0 {
		return nil
	}
	return ValidationError{Errs: append([]error(nil), e...)}
}

func Required(s, fieldName string) error {
	if s == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}

func LengthRange(s, fieldName string, minLen, maxLen int) error {
	n := len(s)
	if minLen >= 0 && n < minLen {
		return errors.New(fmt.Sprintf("%s has %d character%s but must be at least %d", fieldName, n, plural(n), minLen))
	}
	if maxLen >= 0 && n > maxLen {
		return errors.New(fmt.Sprintf("%s has %d character%s but must be at most %d", fieldName, n, plural(n), maxLen))
	}
	return nil
}

func ExactLength(s, fieldName string, n int) error {
	actual := len(s)
	if actual != n {
		return errors.New(fmt.Sprintf("%s has %d character%s but must be exactly %d", fieldName, actual, plural(actual), n))
	}
	return nil
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func Alphanumeric(s, fieldName string, allowSpaces bool) error {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		if allowSpaces && unicode.IsSpace(r) {
			continue
		}
		if allowSpaces {
			return errors.New(fieldName + " must contain only letters, numbers and spaces")
		}
		return errors.New(fieldName + " must contain only letters and numbers")
	}
	return nil
}

func LettersOnly(s, fieldName string) error {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return errors.New(fieldName + " must contain only letters")
		}
	}
	return nil
}

func DigitsOnly(s, fieldName string) error {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return errors.New(fieldName + " must contain only digits")
		}
	}
	return nil
}

func Match(s, fieldName string, pattern *regexp.Regexp) error {
	if !pattern.MatchString(s) {
		return errors.New(fieldName + " must match the required format")
	}
	return nil
}

func DateFormat(s, fieldName string) error {
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return errors.New(fieldName + " must be a valid date (YYYY-MM-DD)")
	}
	return nil
}

var moduleCodePattern = regexp.MustCompile(`^[A-Za-z]{3}[0-9]{3}$`)

func ModuleCodeFormat(s, fieldName string) error {
	if len(s) != 6 {
		return nil
	}
	if !moduleCodePattern.MatchString(s) {
		return errors.New(fieldName + " must be 3 letters followed by 3 digits (e.g. ABC123)")
	}
	return nil
}

func OptionalString(ptr *string, fallback string) string {
	if ptr != nil {
		return *ptr
	}
	return fallback
}

func IntRange(n int, fieldName string, min, max int) error {
	if n < min || n > max {
		return errors.New(fmt.Sprintf("%s must be between %d and %d", fieldName, min, max))
	}
	return nil
}

func DateTimeFormat(s, fieldName string) error {
	if _, err := parseDateTime(s); err != nil {
		return errors.New(fieldName + " must be a valid datetime (RFC3339, e.g. 2006-01-02T15:04:05)")
	}
	return nil
}

func DateRange(startDate, endDate, fieldName string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil
	}
	if end.Before(start) {
		return errors.New(fieldName + " must be on or after startDate")
	}
	return nil
}

func OneOf(s, fieldName string, allowed []string) error {
	for _, a := range allowed {
		if s == a {
			return nil
		}
	}
	return errors.New(fieldName + " must be one of: " + strings.Join(allowed, ", "))
}

func Recurrence(s, fieldName string) error {
	return OneOf(s, fieldName, []string{"daily", "weekly", "biweekly", "monthly"})
}


func parseDateTime(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid datetime")
}

func DateTimeRange(startsAt, endsAt, fieldName string) error {
	start, err := parseDateTime(startsAt)
	if err != nil {
		return nil
	}
	end, err := parseDateTime(endsAt)
	if err != nil {
		return nil
	}
	if !end.After(start) {
		return errors.New(fieldName + " must be after startsAt")
	}
	return nil
}
