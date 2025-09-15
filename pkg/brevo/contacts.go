package brevo

import (
	"context"
	"fmt"

	brevo "github.com/getbrevo/brevo-go/lib"
)

// ContactAttributes represents the attributes for a contact
type ContactAttributes map[string]any

// CreateOrUpdateContactRequest represents the request to create or update a contact
type CreateOrUpdateContactRequest struct {
	Email      string            `json:"email"`
	Attributes ContactAttributes `json:"attributes,omitempty"`
	ListIDs    []int64           `json:"listIds,omitempty"`
}

// CreateOrUpdateContactResponse represents the response from creating or updating a contact
type CreateOrUpdateContactResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

// CreateOrUpdateContact creates a new contact or updates an existing one in Brevo.
func (s *Service) CreateOrUpdateContact(ctx context.Context, req CreateOrUpdateContactRequest) (*CreateOrUpdateContactResponse, error) {
	s.log.Info(ctx, "Creating or updating contact in Brevo", "email", req.Email)
	// Prepare the contact data
	contactData := brevo.CreateContact{
		Email:         req.Email,
		Attributes:    req.Attributes,
		ListIds:       req.ListIDs,
		UpdateEnabled: true,
	}

	// Call the Brevo API
	result, response, err := s.client.ContactsApi.CreateContact(ctx, contactData)
	if err != nil {
		s.log.Error(ctx, "Failed to create/update contact in Brevo",
			"error", err,
			"email", req.Email,
			"response_status", response.Status)
		return nil, fmt.Errorf("brevo: failed to create/update contact: %w", err)
	}

	s.log.Info(ctx, "Successfully created/updated contact in Brevo",
		"email", req.Email,
		"contact_id", result.Id)

	return &CreateOrUpdateContactResponse{
		ID:    result.Id,
		Email: req.Email,
	}, nil
}

// AddContactToListRequest represents the request to add a contact to a list
type AddContactToListRequest struct {
	ContactEmails []string `json:"emails"`
}

// AddContactsToList adds existing contacts to a specific list in Brevo.
func (s *Service) AddContactsToList(ctx context.Context, listID int64, req AddContactToListRequest) error {
	s.log.Info(ctx, "Adding contacts to list in Brevo",
		"list_id", listID,
		"contact_count", len(req.ContactEmails))

	if len(req.ContactEmails) == 0 {
		return fmt.Errorf("brevo: no contact emails provided")
	}

	// Prepare the request data
	addContactData := brevo.AddContactToList{
		Emails: req.ContactEmails,
	}

	// Call the Brevo API
	result, response, err := s.client.ContactsApi.AddContactToList(ctx, listID, addContactData)
	if err != nil {
		s.log.Error(ctx, "Failed to add contacts to list in Brevo",
			"error", err,
			"list_id", listID,
			"contact_emails", req.ContactEmails,
			"response_status", response.Status)
		return fmt.Errorf("brevo: failed to add contacts to list %d: %w", listID, err)
	}

	s.log.Info(ctx, "Successfully added contacts to list in Brevo",
		"list_id", listID,
		"contact_count", len(req.ContactEmails),
		"result", result)

	return nil
}

// DeleteContactRequest represents the request to delete a contact
type DeleteContactRequest struct {
	Email string `json:"email"`
}

// DeleteContact deletes a contact from Brevo.
func (s *Service) DeleteContact(ctx context.Context, req DeleteContactRequest) error {
	s.log.Info(ctx, "Deleting contact in Brevo", "email", req.Email)

	if req.Email == "" {
		return fmt.Errorf("brevo: email is required")
	}

	// Call the Brevo API
	response, err := s.client.ContactsApi.DeleteContact(ctx, req.Email)
	if err != nil {
		s.log.Error(ctx, "Failed to delete contact in Brevo",
			"error", err,
			"email", req.Email,
			"response_status", response.Status)
		return fmt.Errorf("brevo: failed to delete contact: %w", err)
	}

	s.log.Info(ctx, "Successfully deleted contact in Brevo", "email", req.Email)
	return nil
}
