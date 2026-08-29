package usersrepository

import (
	"context"
	"errors"
	"fmt"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/jackc/pgx/v5"
)

func (r *repo) CreateVerificationToken(
	ctx context.Context,
	input usersdomain.NewVerificationToken,
) (usersdomain.VerificationToken, error) {
	if input.MaximumIssues <= 0 {
		return usersdomain.VerificationToken{}, errors.New("verification token issue limit must be positive")
	}
	if len(input.TokenDigest) != 32 || input.TokenKeyID == "" || input.TokenVersion <= 0 {
		return usersdomain.VerificationToken{}, errors.New("verification token digest metadata is invalid")
	}
	maximumIssues, err := safecast.Int32(input.MaximumIssues)
	if err != nil {
		return usersdomain.VerificationToken{}, fmt.Errorf("validate verification token issue limit: %w", err)
	}
	tokenType, err := verificationTokenType(input.TokenType)
	if err != nil {
		return usersdomain.VerificationToken{}, err
	}

	var created usersdomain.VerificationToken
	err = r.withinTransaction(ctx, func(queries usersql.Querier) error {
		if err := queries.AcquireVerificationTokenIssueLock(
			ctx,
			usersql.AcquireVerificationTokenIssueLockParams{
				Email:     input.Email,
				TokenType: input.TokenType,
			},
		); err != nil {
			return fmt.Errorf("lock verification token issuance: %w", err)
		}

		issueCount, err := queries.CountRecentVerificationTokenIssues(
			ctx,
			usersql.CountRecentVerificationTokenIssuesParams{
				Email:          input.Email,
				TokenType:      tokenType,
				RateLimitSince: input.RateLimitSince,
			},
		)
		if err != nil {
			return fmt.Errorf("count recent verification token issues: %w", err)
		}
		if issueCount >= maximumIssues {
			return usersdomain.ErrTooManyAttempts
		}

		row, err := queries.CreateVerificationToken(ctx, usersql.CreateVerificationTokenParams{
			Email:        input.Email,
			ExpiresAt:    input.ExpiresAt,
			TokenType:    tokenType,
			TokenDigest:  append([]byte(nil), input.TokenDigest...),
			TokenKeyID:   input.TokenKeyID,
			TokenVersion: input.TokenVersion,
			IssuedAt:     input.IssuedAt,
		})
		if err != nil {
			if platformdatabase.Classify(err) == platformdatabase.ErrorClassUniqueViolation {
				return usersdomain.ErrTokenCollision
			}
			return fmt.Errorf("create verification token: %w", err)
		}
		created = mapCreatedVerificationToken(row)
		return nil
	})
	if err != nil {
		return usersdomain.VerificationToken{}, err
	}
	return created, nil
}

func (r *repo) ConsumeVerificationToken(
	ctx context.Context,
	input usersdomain.ConsumeVerificationToken,
) (usersdomain.VerificationToken, error) {
	if len(input.TokenTypes) == 0 || len(input.TokenDigests) == 0 ||
		len(input.TokenDigests) != len(input.TokenKeyIDs) ||
		len(input.TokenDigests) != len(input.TokenVersions) {
		return usersdomain.VerificationToken{}, usersdomain.ErrInvalidToken
	}

	tokenTypes := make([]string, len(input.TokenTypes))
	for index, inputType := range input.TokenTypes {
		if _, err := verificationTokenType(inputType); err != nil {
			return usersdomain.VerificationToken{}, usersdomain.ErrInvalidToken
		}
		tokenTypes[index] = inputType
	}

	row, err := r.queries.ConsumeVerificationToken(ctx, usersql.ConsumeVerificationTokenParams{
		ConsumedAt:    input.ConsumedAt,
		Email:         input.Email,
		TokenTypes:    tokenTypes,
		TokenDigests:  cloneBytes(input.TokenDigests),
		TokenKeyIds:   append([]string(nil), input.TokenKeyIDs...),
		TokenVersions: append([]int16(nil), input.TokenVersions...),
		LegacyToken:   input.LegacyToken,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usersdomain.VerificationToken{}, usersdomain.ErrInvalidToken
		}
		return usersdomain.VerificationToken{}, fmt.Errorf("consume verification token: %w", err)
	}
	return mapConsumedVerificationToken(row), nil
}

func (r *repo) InvalidateTokens(ctx context.Context, email string) error {
	if _, err := r.queries.InvalidateVerificationTokens(
		ctx,
		usersql.InvalidateVerificationTokensParams{Email: email},
	); err != nil {
		return fmt.Errorf("invalidate verification tokens: %w", err)
	}
	return nil
}

func verificationTokenType(value string) (usersql.TokenType, error) {
	tokenType := usersql.TokenType(value)
	if !tokenType.Valid() {
		return "", fmt.Errorf("unsupported verification token purpose %q", value)
	}
	return tokenType, nil
}

func cloneBytes(values [][]byte) [][]byte {
	result := make([][]byte, len(values))
	for index, value := range values {
		result[index] = append([]byte(nil), value...)
	}
	return result
}

func mapCreatedVerificationToken(row usersql.CreateVerificationTokenRow) usersdomain.VerificationToken {
	return usersdomain.VerificationToken{
		ID:           row.ID,
		Email:        row.Email,
		UserID:       row.UserID,
		ExpiresAt:    row.ExpiresAt,
		UsedAt:       row.UsedAt,
		TokenType:    row.TokenType,
		TokenKeyID:   derefString(row.TokenKeyID),
		TokenVersion: derefInt16(row.TokenVersion),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func mapConsumedVerificationToken(row usersql.ConsumeVerificationTokenRow) usersdomain.VerificationToken {
	return usersdomain.VerificationToken{
		ID:           row.ID,
		Email:        row.Email,
		UserID:       row.UserID,
		ExpiresAt:    row.ExpiresAt,
		UsedAt:       row.UsedAt,
		TokenType:    row.TokenType,
		TokenKeyID:   derefString(row.TokenKeyID),
		TokenVersion: derefInt16(row.TokenVersion),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func derefInt16(value *int16) int16 {
	if value == nil {
		return 0
	}
	return *value
}
