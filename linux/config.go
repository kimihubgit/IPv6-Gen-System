package main

import (
	"encoding/json"
	"os"
	"time"
)

const StateFileName = "assigned_ips.json"

// AssignedIPRecord tracks batches of IPs assigned to interfaces
type AssignedIPRecord struct {
	Interface string    `json:"interface"`
	Prefix    string    `json:"prefix"`
	Store     string    `json:"store"`
	CreatedAt time.Time `json:"created_at"`
	IPs       []string  `json:"ips"`
}

// LoadAssignedRecords loads previously saved records from disk
func LoadAssignedRecords() ([]AssignedIPRecord, error) {
	data, err := os.ReadFile(StateFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return []AssignedIPRecord{}, nil
		}
		return nil, err
	}

	var records []AssignedIPRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return []AssignedIPRecord{}, nil
	}
	return records, nil
}

// SaveAssignedRecord appends or updates a record and saves to disk
func SaveAssignedRecord(record AssignedIPRecord) error {
	records, _ := LoadAssignedRecords()
	records = append(records, record)
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StateFileName, data, 0644)
}

// ClearAllAssignedRecords clears the records file
func ClearAllAssignedRecords() error {
	return os.Remove(StateFileName)
}
