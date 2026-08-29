package github

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

func (s *Service) safeUserLinkReturnTo(returnTo string) (string, error) {
	if strings.TrimSpace(returnTo) == "" {
		return "", errors.New("return path is required")
	}
	parsed, err := url.Parse(returnTo)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() || parsed.Host != "" {
		if parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return "", errors.New("return URL scheme is not allowed")
		}
		if !s.isAllowedUserLinkReturnURL(parsed) {
			return "", errors.New("return URL host is not allowed")
		}
		return parsed.String(), nil
	}
	if !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "", errors.New("return path must be a relative application path")
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.RequestURI(), nil
}

func (s *Service) isAllowedUserLinkReturnURL(returnURL *url.URL) bool {
	if returnURL == nil {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(returnURL.Hostname()))
	if host == "" {
		return false
	}
	website, err := url.Parse(s.cfg.WebsiteURL)
	if err != nil {
		return false
	}
	configuredHost := strings.ToLower(strings.TrimSpace(website.Hostname()))
	if configuredHost == "" || !strings.EqualFold(returnURL.Scheme, website.Scheme) {
		return false
	}
	if isLocalWebsiteHost(configuredHost) {
		return host == configuredHost
	}
	return host == configuredHost || strings.HasSuffix(host, "."+configuredHost)
}

type webhookEnvelope struct {
	Action     string `json:"action"`
	RefType    string `json:"ref_type"`
	Repository struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		HTMLURL       string `json:"html_url"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
	Sender struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	} `json:"sender"`
	Issue struct {
		ID      int64  `json:"id"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
		User    struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
		Assignee *struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"assignee"`
		Labels []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"labels"`
	} `json:"issue"`
	PullRequest struct {
		ID      int64  `json:"id"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
		Draft   bool   `json:"draft"`
		Merged  bool   `json:"merged"`
		Head    struct {
			Ref     string `json:"ref"`
			HTMLURL string `json:"html_url"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
		User struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
		Assignee *struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"assignee"`
		Labels []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"labels"`
	} `json:"pull_request"`
	Review struct {
		ID    int64  `json:"id"`
		State string `json:"state"`
		Body  string `json:"body"`
		User  struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
		HTMLURL string `json:"html_url"`
	} `json:"review"`
	Comment struct {
		ID        int64     `json:"id"`
		Body      string    `json:"body"`
		HTMLURL   string    `json:"html_url"`
		CreatedAt time.Time `json:"created_at"`
		User      struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
	} `json:"comment"`
	CheckRun struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Status       string `json:"status"`
		Conclusion   string `json:"conclusion"`
		HTMLURL      string `json:"html_url"`
		PullRequests []struct {
			ID     int64 `json:"id"`
			Number int   `json:"number"`
		} `json:"pull_requests"`
	} `json:"check_run"`
	Label struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"label"`
	Assignee *struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	} `json:"assignee"`
	Ref     string `json:"ref"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
	} `json:"commits"`
}
