package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/XigmaDev/huawei"
)

func calculateSMSParts(charCount int) int {
	if charCount <= 70 {
		return 1 // First page: 70 characters
	} else if charCount <= 70+64 {
		return 2 // Second page: 64 characters (total 134)
	} else if charCount <= 70+64+67 {
		return 3 // Third page: 67 characters (total 201)
	} else {
		// After third page, each additional page is 67 characters
		remainingChars := charCount - (70 + 64)       // Subtract first two pages
		additionalPages := (remainingChars + 66) / 67 // Ceiling division for pages of 67
		return 2 + additionalPages                    // Add the first two pages
	}
}

func main() {
	// Read the message from message.txt
	message, err := os.ReadFile("message.txt")
	if err != nil {
		fmt.Println("Error reading message.txt:", err)
		return
	}
	msg := string(message)
	// Count the number of characters (not bytes) in the message
	charCount := utf8.RuneCountInString(msg)
	fmt.Println("Message Character: ", charCount)
	// Read phone numbers from number.txt
	file, err := os.Open("number.txt")
	if err != nil {
		fmt.Println("Error opening number.txt:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var phoneNumbers []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" { // Skip empty lines
			phoneNumbers = append(phoneNumbers, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading number.txt:", err)
		return
	}
	numRecipients := len(phoneNumbers)

	smsPartsPerMessage := calculateSMSParts(charCount)
	fmt.Println("SMS PART : ", smsPartsPerMessage)
	// Calculate total SMS to be sent
	totalSMS := numRecipients * smsPartsPerMessage

	// Check if total SMS is under 500
	if totalSMS >= 500 {
		fmt.Printf("Total SMS to be sent: %d, which is not under 500. Aborting.\n", totalSMS)
		return
	}

	// Login to Huawei modem
	h := huawei.NewHuawei("http://192.168.8.1")
	if err := h.Login("admin", "admin"); err != nil {
		fmt.Println("Login failed:", err)
		return
	}

	// Send messages to all phone numbers
	for _, phone := range phoneNumbers {
		h.SendSMS(msg, phone)
		time.Sleep(3 * time.Second)
	}
}
