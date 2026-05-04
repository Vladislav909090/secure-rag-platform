package model

import "time"

// Document — логическая сущность документа.
type Document struct {
	ID             int64
	UUID           string
	Title          string
	Description    *string
	Attributes     map[string]any
	FileName       string
	FileExtension  string
	MimeType       string
	SizeBytes      int64
	ChecksumSHA256 string
	StorageKey     string
	IndexStatus    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

const (
	IndexStatusPending    = "PENDING"
	IndexStatusProcessing = "PROCESSING"
	IndexStatusReady      = "READY"
	IndexStatusFailed     = "FAILED"
)
