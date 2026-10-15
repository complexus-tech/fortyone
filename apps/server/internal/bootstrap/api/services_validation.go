package api

import "fmt"

func (s services) validate() error {
	if s.workAutomations == nil {
		return fmt.Errorf("work automation service is required")
	}
	if s.workPresets == nil {
		return fmt.Errorf("missing service: work presets")
	}
	if s.activities == nil {
		return fmt.Errorf("missing service: activities")
	}
	if s.admin == nil {
		return fmt.Errorf("missing service: admin")
	}
	if s.attachments == nil {
		return fmt.Errorf("missing service: attachments")
	}
	if s.calendar == nil {
		return fmt.Errorf("missing service: calendar")
	}
	if s.chatSessions == nil {
		return fmt.Errorf("missing service: chatSessions")
	}
	if s.comments == nil {
		return fmt.Errorf("missing service: comments")
	}
	if s.documents == nil {
		return fmt.Errorf("missing service: documents")
	}
	if s.desktopOperations == nil {
		return fmt.Errorf("missing service: desktopOperations")
	}
	if s.developerCredentials == nil {
		return fmt.Errorf("missing service: developerCredentials")
	}
	if s.developerOAuth == nil {
		return fmt.Errorf("missing service: developerOAuth")
	}
	if s.developerOAuthApps == nil {
		return fmt.Errorf("missing service: developerOAuthApps")
	}
	if s.developerAccess == nil {
		return fmt.Errorf("missing service: developerAccess")
	}
	if s.idempotency == nil {
		return fmt.Errorf("missing service: idempotency")
	}
	if s.outboundWebhooks == nil {
		return fmt.Errorf("missing service: outboundWebhooks")
	}
	if s.emailReply == nil {
		return fmt.Errorf("missing service: emailReply")
	}
	if s.epics == nil {
		return fmt.Errorf("missing service: epics")
	}
	if s.feedback == nil {
		return fmt.Errorf("missing service: feedback")
	}
	if s.figma == nil {
		return fmt.Errorf("missing service: figma")
	}
	if s.googleDrive == nil {
		return fmt.Errorf("missing service: googleDrive")
	}
	if s.github == nil {
		return fmt.Errorf("missing service: github")
	}
	if s.slack == nil {
		return fmt.Errorf("missing service: slack")
	}
	if s.integrationRequests == nil {
		return fmt.Errorf("missing service: integrationRequests")
	}
	if s.invitations == nil {
		return fmt.Errorf("missing service: invitations")
	}
	if s.keyResults == nil {
		return fmt.Errorf("missing service: keyResults")
	}
	if s.labels == nil {
		return fmt.Errorf("missing service: labels")
	}
	if s.customFields == nil {
		return fmt.Errorf("missing service: customFields")
	}
	if s.dataExport == nil {
		return fmt.Errorf("missing service: dataExport")
	}
	if s.workspaceSecurity == nil {
		return fmt.Errorf("missing service: workspaceSecurity")
	}
	if s.scim == nil {
		return fmt.Errorf("missing service: scim")
	}
	if s.enterpriseSSO == nil {
		return fmt.Errorf("missing service: enterpriseSSO")
	}
	if s.links == nil {
		return fmt.Errorf("missing service: links")
	}
	if s.maya == nil {
		return fmt.Errorf("missing service: maya")
	}
	if s.notifications == nil {
		return fmt.Errorf("missing service: notifications")
	}
	if s.objectives == nil {
		return fmt.Errorf("missing service: objectives")
	}
	if s.objectiveStats == nil {
		return fmt.Errorf("missing service: objectiveStats")
	}
	if s.okrActivities == nil {
		return fmt.Errorf("missing service: okrActivities")
	}
	if s.reports == nil {
		return fmt.Errorf("missing service: reports")
	}
	if s.search == nil {
		return fmt.Errorf("missing service: search")
	}
	if s.sprints == nil {
		return fmt.Errorf("missing service: sprints")
	}
	if s.states == nil {
		return fmt.Errorf("missing service: states")
	}
	if s.stories == nil {
		return fmt.Errorf("missing service: stories")
	}
	if s.subscriptions == nil {
		return fmt.Errorf("missing service: subscriptions")
	}
	if s.teams == nil {
		return fmt.Errorf("missing service: teams")
	}
	if s.teamSettings == nil {
		return fmt.Errorf("missing service: teamSettings")
	}
	if s.users == nil {
		return fmt.Errorf("missing service: users")
	}
	if s.workspaces == nil {
		return fmt.Errorf("missing service: workspaces")
	}

	return nil
}
