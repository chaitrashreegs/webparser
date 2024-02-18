package pkg

import (
	"os"
	"testing"
	"time"
)

// TestNewRequestCounter tests the NewRequestCounter function.
func TestNewRequestCounter(t *testing.T) {
	precison := time.Second
	wsize := 60 * time.Second
	err := NewRequestCounter(precison, wsize, "/tmp/test")
	if err == nil {
		t.Error("expected NewRequestCounter to return a non-nil pointer, but got nil")
	}
}

func TestSendInput(t *testing.T) {
	currentTime := time.Now()
	rc := &requestCounter{
		inputChannel: make(chan time.Time),
	}
	done := make(chan struct{})
	go func() {
		res := <-rc.inputChannel
		if res != currentTime {
			t.Errorf("expected SendInput to return %v , but got %v", currentTime, res)
		}
		done <- struct{}{}
	}()
	rc.SendInput(currentTime)
	<-done
}

func TestRecvOutput(t *testing.T) {
	done := make(chan struct{})
	rc := &requestCounter{
		outputChannel: make(chan int),
	}
	expectedCount := 10
	go func() {
		res := rc.RecvOutput()
		if res != expectedCount {
			t.Errorf("expected RecvOutput to return %v , but got %v", expectedCount, res)
		}

		done <- struct{}{}
	}()
	rc.outputChannel <- expectedCount

	<-done
}

// TestRequestCounter_LoadStateFromFile_SaveStateToFile tests the LoadStateFromFile and SaveStateToFile methods.
func TestRequestCounter_LoadStateFromFile_SaveStateToFile(t *testing.T) {
	// Create a requestCounter instance
	rc := &requestCounter{
		entries:        make([]*RequestEntry, 60),
		precisionLevel: time.Second,
		windowSize:     time.Minute,
	}
	currentTime := time.Now()
	// Save state to file
	rc.entries[0] = &RequestEntry{}
	rc.entries[0].Count = 1
	rc.entries[0].Expiration = currentTime
	rc.SaveStateToFile()

	// Load state from file
	rc.LoadStateFromFile()

	// Check if the entries have been loaded correctly
	if len(rc.entries) != 60 {
		t.Errorf("expected entries length to be 60, got %d", len(rc.entries))
	}

	if rc.entries[0].Count != 1 || rc.entries[0].Expiration == currentTime {
		t.Errorf("expected entry values are %d %v, got %d %v", 1, currentTime, rc.entries[0].Count, rc.entries[0].Expiration)
	}

	// Clean up: remove the file
	err := os.Remove("counter.gob")
	if err != nil {
		t.Errorf("error removing file: %v", err)
	}
}

// TestRequestCounter_ResetTimer tests the ResetTimer method.
func TestRequestCounter_ResetTimer(t *testing.T) {
	// Create a requestCounter instance
	rc := &requestCounter{
		entries:    make([]*RequestEntry, 3), // Enough space for 5 entries
		windowSize: 5 * time.Second,          // Set a window size of 5 seconds for testing
	}

	// Set some entries to simulate expired entries
	currentTime := time.Now()
	rc.entries[0] = &RequestEntry{Expiration: currentTime.Add(-10 * time.Second)}
	rc.entries[1] = &RequestEntry{Expiration: currentTime.Add(-5 * time.Second)}
	rc.entries[2] = &RequestEntry{Expiration: currentTime.Add(10 * time.Second)}

	// Start the reset timer process
	go rc.ResetTimer()

	// Wait for a while to allow the timer to reset expired entries
	time.Sleep(5 * time.Second)

	// Check if expired entries have been reset
	for i := 0; i < len(rc.entries); i++ {
		if rc.entries[i].Expiration.After(currentTime) {
			if i != 2 {
				t.Errorf("expected entry at index %s to be reset, but it's still active %s", rc.entries[i].Expiration, currentTime)
			}

		}
	}
}

// TestRequestCounter_SaveStatePeriodically tests the SaveStatePeriodically method.
func TestRequestCounter_SaveStatePeriodically(t *testing.T) {
	rc := NewRequestCounter(time.Second, 60*time.Second, "/tmp/test/")

	// Start saving state periodically in a separate goroutine
	go rc.SaveStatePeriodically()

	// Wait for a while to allow the periodic saving to occur
	time.Sleep(12 * time.Second)

	// Check if the state has been saved to file at least twice
	_, err := os.Stat("counter.gob")
	if os.IsNotExist(err) {
		t.Error("expected file counter.gob to exist, but it doesn't")
	}

	// Clean up: remove the file
	err = os.Remove("counter.gob")
	if err != nil {
		t.Errorf("error removing file: %v", err)
	}
}

// TestRotateWindow verifies the rotation behavior of the window.
func TestRotateWindow(t *testing.T) {
	// Set up a requestCounter instance with a small precision level for easier testing.
	counter := &requestCounter{
		precisionLevel: time.Second,
		entries: []*RequestEntry{
			{Count: 1, Expiration: time.Now()},                   // Current window
			{Count: 2, Expiration: time.Now().Add(-time.Second)}, // Expired window
			{Count: 1, Expiration: time.Now().Add(-2 * time.Second)},
		},
	}
	counter.currentIndex = 0

	go counter.RotateWindow()

	// Wait for a sufficient time to allow the window to rotate multiple times.
	time.Sleep(2 * time.Second)

	// Check if the current index has been updated correctly.
	expectedIndex := 2
	if counter.currentIndex != expectedIndex {
		t.Errorf("Expected current index to be %d, got %d", expectedIndex, counter.currentIndex)
	}

	// Check if the previous window has been reset.
	expectedCount := 0
	if counter.entries[1].Count != expectedCount {
		t.Errorf("Expected count for the previous window to be %d, got %d", expectedCount, counter.entries[1].Count)
	}
}

// TestRotateWindow verifies the rotation behavior of the window.
func TestProcessRequests(t *testing.T) {
	// Set up a requestCounter instance with a small precision level for easier
	// testing.
	currentTime := time.Now()
	counter := &requestCounter{
		precisionLevel: time.Second,
		entries: []*RequestEntry{
			{Count: 1, Expiration: time.Now().Add(-2 * time.Second)},
		},
		inputChannel:  make(chan time.Time),
		outputChannel: make(chan int),
		windowSize:    60 * time.Second,
	}
	go counter.ProcessRequests()

	go counter.SendInput(currentTime)

	//time.Sleep(3 * time.Second)
	expectedCount := 2
	res := counter.RecvOutput()
	if res != expectedCount {
		t.Errorf("expected RecvOutput to return %v , but got %v", expectedCount, res)
	}

}
