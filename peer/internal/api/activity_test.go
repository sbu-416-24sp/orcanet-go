package api

import (
	"fmt"
	"sync"
	"testing"
)

func TestActivityConcurrency(t *testing.T) {
	var wg sync.WaitGroup
	b := NewBackend() // Assuming you have a constructor for Backend which initializes the mutex, activities slice, and anything else required

	// No direct reset to activities is needed here if NewBackend() provides a fresh state

	// Define the number of goroutines to run
	numGoroutines := 50

	// Add activities concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			act := Activity{ID: id, Name: "Activity" + fmt.Sprint(id)}
			err := b.SetActivity(act)
			if err != nil {
				t.Errorf("Error adding activity: %v", err)
			}
		}(i)
	}

	// Assume some delay here if necessary to ensure adds are processed before removes
	// However, in a real concurrent scenario, adds and removes can be intermingled.

	// Remove activities concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := b.RemoveActivity(id)
			if err != nil {
				t.Errorf("Error removing activity: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Final check to see if all activities were added and removed correctly
	finalActivities, err := b.GetActivities()
	if err != nil {
		t.Errorf("Error retrieving activities: %v", err)
	}
	if len(finalActivities) != 0 {
		t.Errorf("Expected 0 activities after adds and removes, got %d", len(finalActivities))
	}
}
