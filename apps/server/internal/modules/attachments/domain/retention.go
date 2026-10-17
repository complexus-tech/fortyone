package attachmentdomain

import "time"

type PendingChatUploadRetention struct {
	CreatedBefore   time.Time
	RetiredAt       time.Time
	Limit           int
	StorageProvider string
	ContainerName   string
}
