package pkg

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const fileNameToSave = "counter.gob"

// RequestEntry corresponds to single entry in the request counter.
type RequestEntry struct {
	Count      int       // Count of requests
	Expiration time.Time // Expiration time for the count value
}

// requestCounter stores the request counts and Expiration times for the last 60 seconds.
type requestCounter struct {
	entries       []*RequestEntry
	inputChannel  chan time.Time // Channel to signal arrival of requests
	outputChannel chan int       // return total count in of window

	shutdownCh     chan struct{} // Channel to signal graceful shutdown
	precisionLevel time.Duration // Precision level for calculating window index
	windowSize     time.Duration // Number of time windows to keep track of
	currentIndex   int
	filePath       string
	sync.Mutex
}

// NewRequestCounter creates a new requestCounter with the given precision and
// window size.
func NewRequestCounter(precison, wsize time.Duration, filePath string) Parser {
	r := &requestCounter{}
	r.inputChannel = make(chan time.Time)
	r.outputChannel = make(chan int)
	r.shutdownCh = make(chan struct{})
	r.precisionLevel = precison
	r.windowSize = wsize
	r.currentIndex = 0
	bufferSize := wsize.Nanoseconds() / precison.Nanoseconds()
	r.entries = make([]*RequestEntry, bufferSize)
	r.filePath = fmt.Sprintf("%s/%s", filePath, fileNameToSave)
	return r
}

// Parser holds the methods to be implemented by the requestCounter.
type Parser interface {
	GetWindowSize() time.Duration
	LoadStateFromFile()
	ProcessRequests()
	SaveStatePeriodically()
	ResetTimer()
	SendInput(time.Time)
	RecvOutput() int
	HandleShutdown()
	SaveStateToFile()
	RotateWindow()
}

// GetWindowSize returns the window size of the requestCounter.
func (r *requestCounter) GetWindowSize() time.Duration {
	return r.windowSize
}

// SendInput sends the input time to the requestCounter.
func (r *requestCounter) SendInput(input time.Time) {
	r.inputChannel <- input
}

// RecvOutput returns the total count of requests from output Channel.
func (r *requestCounter) RecvOutput() int {
	return <-r.outputChannel
}

// Initializaion loads the state from file, and starts the goroutines for
// saving state, resetting timer and rotating window and processing requests.
func Initializaion(c Parser) {
	c.LoadStateFromFile()
	go c.SaveStatePeriodically()
	go c.ResetTimer()
	go c.RotateWindow()
	go c.ProcessRequests()

}

// RotateWindow rotates the window every precision level and resets the count
// for the new window.
func (r *requestCounter) RotateWindow() {
	t := time.NewTicker(r.precisionLevel)
	for {
		select {
		case <-t.C:
			r.Lock()
			r.currentIndex++
			if r.currentIndex == len(r.entries) {
				r.currentIndex = 0
			}
			r.entries[r.currentIndex] = &RequestEntry{}
			r.Unlock()
		case <-r.shutdownCh:
			return
		}

	}
}

// LoadStateFromFile loads the state from file if it exists.
func (r *requestCounter) LoadStateFromFile() {
	f, err := os.Open(r.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal("Error opening file:", err)
		}
		return
	}

	defer f.Close()
	d := []RequestEntry{}
	decoder := gob.NewDecoder(f)
	err = decoder.Decode(&d)
	if err != nil {
		log.Fatal("Error decoding counter:", err)
	}
	currentTime := time.Now().Add(-1 * r.windowSize)
	j := 0
	for _, entry := range d {
		if entry.Expiration.After(currentTime) {
			r.entries[j] = &RequestEntry{
				Count:      entry.Count,
				Expiration: entry.Expiration,
			}
			r.currentIndex = j
			j++
		}
	}
}

// ProcessRequests processes the requests and returns the total count of
// requests in the last window size.
func (r *requestCounter) ProcessRequests() {
	for {
		select {
		case input := <-r.inputChannel:
			r.Lock()
			r.entries[r.currentIndex].Count++
			r.entries[r.currentIndex].Expiration = input.Add(r.windowSize)
			totalCounts := 0
			for _, entry := range r.entries {
				// check expiration time was before input-windowsize
				if entry != nil {
					if entry.Expiration.After(input) {
						totalCounts = totalCounts + entry.Count
					}
				}
			}
			r.outputChannel <- totalCounts
			r.Unlock()
		case <-r.shutdownCh:
			return
		}
	}
}

// SaveStatePeriodically saves the state to file every 10 seconds.
func (r *requestCounter) SaveStatePeriodically() {
	ticker := time.NewTicker(10 * time.Second) // every 10 seconds counter value will be persisted on to file
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// Save state immediately if signaled
			r.SaveStateToFile()
		case <-r.shutdownCh:
			return
		}
	}
}

// HandleShutdown handles the shutdown signal and saves the state to file.
func (r *requestCounter) HandleShutdown() {
	//gracefully shutting down , save the counter value to file
	r.SaveStateToFile()
	log.Printf("saved file to %v", r.filePath)
	r.shutdownCh <- struct{}{}
}

// SaveStateToFile saves the state to file.
func (r *requestCounter) SaveStateToFile() {
	f, err := os.Create(r.filePath)
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer f.Close()
	r.Lock()
	d := make([]RequestEntry, 0)
	for _, v := range r.entries {
		if v != nil {
			d = append(d, *v)
		}
	}
	r.Unlock()
	encoder := gob.NewEncoder(f)
	err = encoder.Encode(&d)
	if err != nil {
		log.Fatal("Error encoding counter:", err)
	}
}

// ResetTimer resets the timer and checks for expired entries every window size.
func (r *requestCounter) ResetTimer() {
	ticker := time.NewTicker(r.windowSize) // Check for expired entries on each window size interval
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.Lock()
			currentTime := time.Now()
			for i, entry := range r.entries {
				if entry.Expiration.Before(currentTime) {
					r.entries[i] = &RequestEntry{}
				}
			}
			r.Unlock()
		case <-r.shutdownCh:
			return
		}
	}
}
