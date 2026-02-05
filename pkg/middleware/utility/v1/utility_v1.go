package utilityV1

import (
	"fmt"
	"log"
	"math/rand"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func GetEnv(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		fmt.Println("Error loading .env file")
		log.Fatalf("Error loading .env file")
		return err.Error()
	}
	return os.Getenv(key)
}

func IsNumeric(input string) bool {
	pattern := "^[0-9]+(\\.[0-9]*)?$"
	match, err := regexp.MatchString(pattern, input)
	return err == nil && match
}

func HasAlphabetsAndWhitespace(input string) bool {
	pattern := "^[a-zA-Z\\s]+$"
	match, err := regexp.MatchString(pattern, input)
	return err == nil && match
}

func IsEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Generate sequencial number
func GenerateSequenceNumber(max_digits, current_count int) string {
	max_digits = max_digits - 1
	var instructionID string
	current_length := len(strconv.Itoa(current_count))

	if current_length <= max_digits {
		current_count++
		for strL := 0; strL <= max_digits-current_length; strL++ {
			instructionID += "0"
		}
	} else {
		current_count = 1
		for strL := 0; strL <= max_digits-current_length; strL++ {
			instructionID += "0"
		}
	}

	instructionID += strconv.Itoa(current_count)
	return instructionID
}

// Invalid password, should have at least 8 characters long, a mix of uppercase and lowercase letters and at least one special character (@ or .)
// Validate password
func IsPasswordValid(password string) bool {
	hasEightLen := false
	hasUpperChar := false
	hasLowerChar := false
	hasSpecialChar := false
	if len(password) >= 8 {
		hasEightLen = true
	}

	upperString := regexp.MustCompile(`[A-Z]`)
	lowerString := regexp.MustCompile(`[a-z]`)
	specialString := regexp.MustCompile(`[!@#$%^&*(.)]`)

	hasUpperChar = upperString.MatchString(password)
	hasLowerChar = lowerString.MatchString(password)
	hasSpecialChar = specialString.MatchString(password)

	return hasEightLen && hasUpperChar && hasLowerChar && hasSpecialChar
}

func GenerateRandomStrings(maxLen int, letterType []string) string {
	var prefix, letterBytes string
	for _, typeValue := range letterType {
		typeValue = strings.ToUpper(typeValue)
		switch typeValue {
		case "UPPERCASE":
			letterBytes += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		case "LOWERCASE":
			letterBytes += "abcdefghijklmnopqrstuvwxyz"
		case "NUMERIC":
			letterBytes += "1234567890"
		default:
			letterBytes += "Invalid letter type"
		}
	}

	if letterBytes != "Invalid letter type" {
		source := rand.NewSource(time.Now().UnixNano())
		random := rand.New(source)

		for maxLen > 0 {
			prefix += string(letterBytes[random.Intn(len(letterBytes))])
			maxLen--
		}
		return prefix
	}

	return letterBytes
}

// Hide text in console
func HidePassword(password string) string {
	passLen := len(password)
	newLen := passLen - 3
	tempPass := ""
	for newLen >= 0 {
		tempPass += "x"
		newLen--
	}
	tempPass += password[passLen-3 : passLen]
	return tempPass
}
