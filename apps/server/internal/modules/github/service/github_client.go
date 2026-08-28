package github

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	githubsdk "github.com/google/go-github/v72/github"
	"github.com/google/uuid"
)

func (s *Service) createAppJWT() (string, error) {
	if s.privateKey == nil || s.cfg.AppID == 0 {
		return "", errors.New("github app credentials are not configured")
	}
	now := s.currentTime()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Add(-time.Minute).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": s.cfg.AppID,
	})
	return token.SignedString(s.privateKey)
}

func (s *Service) newAppClient() (*githubsdk.Client, error) {
	appJWT, err := s.createAppJWT()
	if err != nil {
		return nil, err
	}
	return githubsdk.NewClient(s.httpClient).WithAuthToken(appJWT), nil
}

func (s *Service) newInstallationClient(ctx context.Context, installationID int64) (*githubsdk.Client, error) {
	token, err := s.getInstallationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}
	return githubsdk.NewClient(s.httpClient).WithAuthToken(token), nil
}

func (s *Service) getInstallationToken(ctx context.Context, installationID int64) (string, error) {
	client, err := s.newAppClient()
	if err != nil {
		return "", err
	}
	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return "", err
	}
	return token.GetToken(), nil
}

func (s *Service) getInstallation(ctx context.Context, installationID int64) (githubInstallationPayload, error) {
	client, err := s.newAppClient()
	if err != nil {
		return githubInstallationPayload{}, err
	}
	installation, _, err := client.Apps.GetInstallation(ctx, installationID)
	if err != nil {
		return githubInstallationPayload{}, err
	}
	return toInstallationPayload(installation), nil
}

func (s *Service) listInstallationRepositories(ctx context.Context, installationID int64) ([]githubRepositoryPayload, error) {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return nil, err
	}

	options := &githubsdk.ListOptions{PerPage: 100}
	items := make([]githubRepositoryPayload, 0)
	for {
		repos, response, err := client.Apps.ListRepos(ctx, options)
		if err != nil {
			return nil, err
		}
		for _, repository := range repos.Repositories {
			if repository == nil {
				continue
			}
			items = append(items, toRepositoryPayload(repository))
		}
		if response == nil || response.NextPage == 0 {
			break
		}
		options.Page = response.NextPage
	}
	return items, nil
}

type githubIssuePayload struct {
	ID      int64  `json:"id"`
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	State   string `json:"state"`
}

func toInstallationPayload(installation *githubsdk.Installation) githubInstallationPayload {
	if installation == nil {
		return githubInstallationPayload{}
	}

	permissions := map[string]string{}
	if installation.Permissions != nil {
		bytes, err := json.Marshal(installation.Permissions)
		if err == nil {
			_ = json.Unmarshal(bytes, &permissions)
		}
	}

	var avatarURL *string
	account := installation.GetAccount()
	accountID := int64(0)
	accountLogin := ""
	accountType := ""
	if account != nil && account.AvatarURL != nil {
		avatarURL = account.AvatarURL
	}
	if account != nil {
		accountID = account.GetID()
		accountLogin = account.GetLogin()
		accountType = account.GetType()
	}

	return githubInstallationPayload{
		ID: installation.GetID(),
		Account: githubInstallationAccountPayload{
			ID:        accountID,
			Login:     accountLogin,
			Type:      accountType,
			AvatarURL: avatarURL,
		},
		RepositorySelection: installation.GetRepositorySelection(),
		Permissions:         permissions,
		Events:              installation.Events,
	}
}

func toRepositoryPayload(repository *githubsdk.Repository) githubRepositoryPayload {
	if repository == nil {
		return githubRepositoryPayload{}
	}

	var description *string
	if repository.Description != nil {
		description = repository.Description
	}

	owner := repository.GetOwner()
	ownerID := int64(0)
	ownerLogin := ""
	if owner != nil {
		ownerID = owner.GetID()
		ownerLogin = owner.GetLogin()
	}

	return githubRepositoryPayload{
		ID:            repository.GetID(),
		Name:          repository.GetName(),
		FullName:      repository.GetFullName(),
		Description:   description,
		HTMLURL:       repository.GetHTMLURL(),
		CloneURL:      repository.GetCloneURL(),
		SSHURL:        repository.GetSSHURL(),
		DefaultBranch: repository.GetDefaultBranch(),
		Private:       repository.GetPrivate(),
		Archived:      repository.GetArchived(),
		Disabled:      repository.GetDisabled(),
		Owner: githubRepositoryOwnerPayload{
			ID:    ownerID,
			Login: ownerLogin,
		},
	}
}

func toIssuePayload(issue *githubsdk.Issue) githubIssuePayload {
	if issue == nil {
		return githubIssuePayload{}
	}
	return githubIssuePayload{
		ID:      issue.GetID(),
		Number:  issue.GetNumber(),
		HTMLURL: issue.GetHTMLURL(),
		Title:   issue.GetTitle(),
		State:   issue.GetState(),
	}
}

func (s *Service) createIssue(ctx context.Context, installationID int64, owner, repository, title, body string) (githubIssuePayload, error) {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return githubIssuePayload{}, err
	}
	request := &githubsdk.IssueRequest{
		Title: &title,
		Body:  &body,
	}
	issue, _, err := client.Issues.Create(ctx, owner, repository, request)
	if err != nil {
		return githubIssuePayload{}, err
	}
	return toIssuePayload(issue), nil
}

func (s *Service) createIssueComment(ctx context.Context, installationID int64, owner, repository string, number int, body string) error {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return err
	}
	comment := &githubsdk.IssueComment{Body: &body}
	_, _, err = client.Issues.CreateComment(ctx, owner, repository, number, comment)
	return err
}

func (s *Service) issueHasStoryComment(ctx context.Context, installationID int64, owner, repository string, number int, storyID uuid.UUID, storyURL string) (bool, error) {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return false, err
	}

	options := &githubsdk.IssueListCommentsOptions{
		ListOptions: githubsdk.ListOptions{PerPage: 100},
	}

	for {
		comments, response, err := client.Issues.ListComments(ctx, owner, repository, number, options)
		if err != nil {
			return false, err
		}

		for _, comment := range comments {
			if comment == nil || comment.Body == nil {
				continue
			}
			body := *comment.Body
			if strings.Contains(body, storyURL) || strings.Contains(body, storyCommentMarker(storyID)) {
				return true, nil
			}
		}

		if response == nil || response.NextPage == 0 {
			return false, nil
		}
		options.Page = response.NextPage
	}
}

func (s *Service) updateIssue(ctx context.Context, installationID int64, owner, repository string, number int, title, body, state string) (githubIssuePayload, error) {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return githubIssuePayload{}, err
	}
	request := &githubsdk.IssueRequest{
		Title: &title,
		Body:  &body,
		State: &state,
	}
	issue, _, err := client.Issues.Edit(ctx, owner, repository, number, request)
	if err != nil {
		return githubIssuePayload{}, err
	}
	return toIssuePayload(issue), nil
}
