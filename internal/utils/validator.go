package utils

import (
	"regexp"
	"time"
)

func IsValidFlightNumber(flightNumber string) bool {
	// Format: AA1234 or AA 1234
	pattern := `^[A-Z]{2}\s?\d{1,4}$`
	matched, _ := regexp.MatchString(pattern, flightNumber)
	return matched
}

func IsValidAirportCode(code string) bool {
	// IATA codes are 3 uppercase letters
	pattern := `^[A-Z]{3}$`
	matched, _ := regexp.MatchString(pattern, code)
	return matched
}

func IsValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func ParseFlightNumber(flightNumber string) (airlineCode string, number string) {
	// Remove spaces and extract airline code and number
	cleaned := regexp.MustCompile(`\s+`).ReplaceAllString(flightNumber, "")
	if len(cleaned) >= 3 {
		airlineCode = cleaned[:2]
		number = cleaned[2:]
	}
	return
}
