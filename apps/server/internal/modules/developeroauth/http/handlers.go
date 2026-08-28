package developeroauthhttp

import (
	"context"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	"github.com/google/uuid"
)

// Service is the transport's complete dependency surface. Keeping this
// interface local prevents HTTP concerns from leaking into the application
// service and makes authorization-boundary tests independent of persistence.
type Service interface {
	CreateManagedApplication(context.Context, developeroauthdomain.ManagementAccess, developeroauth.CreateManagedApplicationInput) (developeroauthdomain.IssuedManagedApplication, error)
	ListManagedApplications(context.Context, developeroauthdomain.ManagementAccess) ([]developeroauthdomain.ManagedApplication, error)
	ListClientSecrets(context.Context, developeroauthdomain.ManagementAccess, uuid.UUID) ([]developeroauthdomain.ClientSecret, error)
	RotateClientSecret(context.Context, developeroauthdomain.ManagementAccess, uuid.UUID, developeroauth.RotateClientSecretInput) (developeroauthdomain.IssuedClientSecret, error)
	RevokeClientSecret(context.Context, developeroauthdomain.ManagementAccess, uuid.UUID, uuid.UUID, developeroauth.RevokeApplicationInput) error
	InstallApplication(context.Context, developeroauthdomain.ManagementAccess, developeroauth.InstallApplicationInput) (developeroauthdomain.ApplicationInstallation, error)
	ListApplicationInstallations(context.Context, developeroauthdomain.ManagementAccess) ([]developeroauthdomain.ApplicationInstallation, error)
	UpdateApplicationInstallation(context.Context, developeroauthdomain.ManagementAccess, uuid.UUID, developeroauth.UpdateApplicationInstallationInput) (developeroauthdomain.ApplicationInstallation, error)
	RevokeApplicationInstallation(context.Context, developeroauthdomain.ManagementAccess, uuid.UUID, developeroauth.RevokeApplicationInput) error
}

var _ Service = (*developeroauth.ApplicationManager)(nil)

type Handlers struct {
	service Service
}

func New(service Service) *Handlers {
	if service == nil {
		panic("developer OAuth application management service is required")
	}
	return &Handlers{service: service}
}
