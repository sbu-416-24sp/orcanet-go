package api

type Activity struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Size         string `json:"size"`
	Hash         string `json:"hash"`
	Status       string `json:"status"`
	ShowDropdown bool   `json:"showDropdown"`
	Peers        int    `json:"peers"`
}

// GetActivities is now safe for concurrent use
func (b *Backend) GetActivities() ([]Activity, error) {
	b.mutex.RLock()         // Lock for reading
	defer b.mutex.RUnlock() // Ensure the lock is released
	return b.activities, nil
}

// SetActivity is now safe for concurrent use
func (b *Backend) SetActivity(activity Activity) error {
	b.mutex.Lock()         // Lock for writing
	defer b.mutex.Unlock() // Ensure the lock is released
	b.activities = append(b.activities, activity)
	return nil
}

// RemoveActivity is now safe for concurrent use
func (b *Backend) RemoveActivity(id int) error {
	b.mutex.Lock()         // Lock for writing
	defer b.mutex.Unlock() // Ensure the lock is released

	for i, activity := range b.activities {
		if activity.ID == id {
			b.activities = append(b.activities[:i], b.activities[i+1:]...)
			return nil
		}
	}
	return nil
}

// UpdateActivityName is now safe for concurrent use
func (b *Backend) UpdateActivityName(id int, name string) error {
	b.mutex.Lock()         // Lock for writing
	defer b.mutex.Unlock() // Ensure the lock is released

	for i, activity := range b.activities {
		if activity.ID == id {
			b.activities[i].Name = name
			return nil
		}
	}
	return nil
}
