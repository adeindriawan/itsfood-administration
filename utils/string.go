package utils

import (
	"errors"
	"regexp"
	"strings"
)

func SanitizePhoneNumber(number string) (string, error) {
	var phoneNumber string = ""
	var errNumberNotValid error = errors.New("nomor telepon tidak valid: nomor telepon tidak mengikuti standar nomor telepon 08xx/628xx")
	var nonAlphanumericRegex = regexp.MustCompile(`\D+`)
	phoneNumber += nonAlphanumericRegex.ReplaceAllString(number, "")

	if len(phoneNumber) > 13 {
		errNumberTooLong := errors.New("nomor telepon terlalu panjang: ada kemungkinan merupakan gabungan dari banyak nomor")
		return "", errNumberTooLong
	}

	if phoneNumber[0:2] == "62" {
		return phoneNumber, nil
	}

	if phoneNumber[0:2] == "08" {
		prefix := "62"
		phoneNumber = prefix + strings.TrimPrefix(phoneNumber, "0")

		return phoneNumber, nil
	}

	return "", errNumberNotValid
}
