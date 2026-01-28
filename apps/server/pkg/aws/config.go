package aws

// Config holds AWS S3-related configuration.
type Config struct {
	AccessKeyID          string
	SecretAccessKey      string
	Region               string
	Endpoint             string
	PublicURL            string
	ForcePathStyle       bool
	ProfileImagesBucket  string
	WorkspaceLogosBucket string
	AttachmentsBucket    string
}
