package azure

// Config holds Azure Storage-related configuration
type Config struct {
	ConnectionString        string
	StorageAccountName      string
	AccountKey              string
	ProfileImagesContainer  string
	WorkspaceLogosContainer string
	AttachmentsContainer    string
}
