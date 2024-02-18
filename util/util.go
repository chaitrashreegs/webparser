package util

import "time"

// UserInput is to store input values from the user
type UserInput struct {
	Address        string
	Port           string
	OutputFilePath string
	Precison       time.Duration
	WindowSize     time.Duration
}
