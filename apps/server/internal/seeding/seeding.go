package seeding

import (
	"context"
	"fmt"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
)

// Seeder manages the database seeding process.
type Seeder struct {
	usersService      *users.Service
	workspacesService *workspaces.Service
}

// NewSeeder creates a new Seeder instance.
func NewSeeder(usersService *users.Service, workspacesService *workspaces.Service) *Seeder {
	return &Seeder{
		usersService:      usersService,
		workspacesService: workspacesService,
	}
}

// SeedData contains the parameters for seeding.
type SeedData struct {
	UserEmail     string
	UserFullName  string
	WorkspaceName string
	WorkspaceSlug string
}

// Run executes the seeding process.
func (s *Seeder) Run(ctx context.Context, data SeedData) error {
	// 1. Check if user exists, if not create them
	user, err := s.usersService.GetUserByEmail(ctx, data.UserEmail)
	if err != nil {
		if err.Error() == "we couldn't find your account" { // users.ErrNotFound
			fmt.Printf("Creating user: %s (%s)\n", data.UserFullName, data.UserEmail)
			user, err = s.usersService.Register(ctx, users.CoreNewUser{
				Email:    data.UserEmail,
				FullName: data.UserFullName,
				Timezone: "UTC",
			})
			if err != nil {
				return fmt.Errorf("failed to register user: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check for existing user: %w", err)
		}
	} else {
		fmt.Printf("User already exists: %s\n", data.UserEmail)
	}

	// 2. Create the workspace
	fmt.Printf("Creating workspace: %s (%s)\n", data.WorkspaceName, data.WorkspaceSlug)

	_, err = s.workspacesService.Create(ctx, workspaces.CoreWorkspace{
		Name:  data.WorkspaceName,
		Slug:  data.WorkspaceSlug,
		Color: "#4f46e5", // Default FortyOne indigo
	}, user.ID)

	if err != nil {
		// If slug is taken, we might want to skip or return error
		if err.Error() == "workspace with this url already exists" { // workspaces.ErrSlugTaken
			fmt.Printf("Workspace with slug '%s' already exists. Skipping creation.\n", data.WorkspaceSlug)
			return nil
		}
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	fmt.Println("Seeding completed successfully!")
	return nil
}
