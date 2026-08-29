package figma

import (
	"context"
	"fmt"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

func (s *Service) RewrapCredentials(
	ctx context.Context,
) (credentialvault.RotationReport, error) {
	if s == nil || s.secrets == nil {
		return credentialvault.RotationReport{}, credentialvault.ErrNotConfigured
	}
	activeKey, err := s.secrets.ActiveKeyRef()
	if err != nil {
		return credentialvault.RotationReport{}, err
	}
	report := credentialvault.RotationReport{ActiveKey: activeKey}
	var after *uuid.UUID
	for {
		records, err := s.repo.ListCredentialsForRewrap(
			ctx,
			after,
			credentialvault.DefaultMaintenanceBatchSize,
		)
		if err != nil {
			return report, fmt.Errorf("list Figma credentials for rewrap: %w", err)
		}
		for _, record := range records {
			report.Scanned++
			result, err := s.secrets.Rewrap(
				figmaprovider.CredentialContext(
					record.WorkspaceID,
					record.ID,
					record.InstallationGeneration,
				),
				record.Payload,
			)
			if err != nil {
				return report, fmt.Errorf("rewrap Figma credential %s: %w", record.ID, err)
			}
			if !result.Changed {
				report.Current++
				continue
			}
			replaced, err := s.repo.RewrapCredential(ctx, record, result.Envelope)
			if err != nil {
				return report, fmt.Errorf("persist Figma credential rewrap %s: %w", record.ID, err)
			}
			if replaced {
				report.Rewrapped++
			} else {
				report.Stale++
			}
		}
		if len(records) < credentialvault.DefaultMaintenanceBatchSize {
			return report, nil
		}
		cursor := records[len(records)-1].ID
		after = &cursor
	}
}
