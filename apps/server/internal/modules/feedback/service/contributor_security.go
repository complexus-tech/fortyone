package feedback

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"math/big"
	"strings"
)

func (s *Service) requireContributorFeatures() error {
	if s.nextRepo == nil || s.security == nil || s.random == nil {
		return ErrFeatureUnavailable
	}
	return nil
}
func (s *Service) randomToken(size int) (string, []byte, error) {
	buffer := make([]byte, size)
	if _, err := io.ReadFull(s.random, buffer); err != nil {
		return "", nil, fmt.Errorf("generate feedback token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buffer)
	return token, hashOpaqueToken(token), nil
}
func (s *Service) randomCode() (string, error) {
	value, err := rand.Int(s.random, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate feedback verification code: %w", err)
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}
func (s *Service) verificationCodeHash(portalID uuid.UUID, email, code string) []byte {
	mac := hmac.New(sha256.New, s.security.codeKey[:])
	_, _ = mac.Write([]byte("fortyone:feedback-verification-code:v1\n"))
	_, _ = mac.Write([]byte(portalID.String()))
	_, _ = mac.Write([]byte("\n" + strings.ToLower(strings.TrimSpace(email)) + "\n" + strings.TrimSpace(code)))
	return mac.Sum(nil)
}
func hashOpaqueToken(token string) []byte {
	hash := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hash[:]
}
func parseAuthorizationToken(value, scheme string) (string, error) {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) != 2 || !strings.EqualFold(parts[0], scheme) || parts[1] == "" {
		return "", ErrContributorSessionInvalid
	}
	return parts[1], nil
}
func normalizeSessionError(err error) error {
	if errors.Is(err, ErrVerificationExpired) || errors.Is(err, ErrVerificationConsumed) || errors.Is(err, ErrVerificationAttempts) || errors.Is(err, ErrContributorBlocked) {
		return err
	}
	if errors.Is(err, ErrNotFound) {
		return ErrContributorSessionInvalid
	}
	return err
}
func normalizeContributorSessionSource(source string) string {
	return strings.ToLower(strings.TrimSpace(source))
}
func maskParticipant(participant CoreParticipant, policy string) CoreParticipant {
	if participant.Kind != ContributorKindVerifiedGuest && participant.Kind != ContributorKindExternal {
		return participant
	}
	if policy == GuestIdentityPolicyAlwaysMaskGuests {
		participant.PublicMasked = true
	}
	return participant
}
