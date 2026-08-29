package feedback

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"github.com/complexus-tech/projects-api/pkg/feedbacksecurity"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"strings"
	"time"
)

const (
	verificationLifetime       = 10 * time.Minute
	contributorSessionLifetime = 30 * 24 * time.Hour
	preferenceSessionLifetime  = 30 * time.Minute
	widgetSessionLifetime      = 24 * time.Hour
	widgetAssertionMaxLifetime = 5 * time.Minute
	widgetClockSkew            = 30 * time.Second
	widgetRotationGrace        = 24 * time.Hour
	maxDisplayNameRunes        = 100
	maxUpdateTitleRunes        = 200
	maxUpdateSummaryRunes      = 500
	maxUpdateBodyRunes         = 50_000
	defaultUpdatesPageSize     = 20
	maxUpdatesPageSize         = 50
	maxMergeCandidatesLimit    = 50
)

type ContributorTasks interface {
	EnqueueFeedbackContributorDelivery(tasks.FeedbackContributorDeliveryPayload) error
}
type contributorSecurity struct {
	codeKey        [sha256.Size]byte
	unsubscribeKey [sha256.Size]byte
	widgetAEAD     cipher.AEAD
	widgetAADBase  []byte
}

func WithContributorFeatures(authSecret, websiteURL string, taskService ContributorTasks) Option {
	return func(service *Service) {
		service.websiteURL = strings.TrimRight(strings.TrimSpace(websiteURL), "/")
		service.tasks = taskService
		service.security = newContributorSecurity(authSecret)
	}
}
func newContributorSecurity(secret string) *contributorSecurity {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil
	}
	codeKey := sha256.Sum256([]byte("fortyone:feedback-verification-code:v1\x00" + secret))
	unsubscribeKey, err := feedbacksecurity.DeriveUnsubscribeKey(secret)
	if err != nil {
		return nil
	}
	widgetKey := sha256.Sum256([]byte("fortyone:feedback-widget-signing-secret:v1\x00" + secret))
	block, err := aes.NewCipher(widgetKey[:])
	if err != nil {
		return nil
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}
	return &contributorSecurity{codeKey: codeKey, unsubscribeKey: unsubscribeKey, widgetAEAD: aead, widgetAADBase: []byte("fortyone:feedback-widget-signing-secret:v1")}
}
