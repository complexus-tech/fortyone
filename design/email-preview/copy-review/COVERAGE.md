# Coverage and implementation notes

This collection covers the 12 file-backed application template types plus notification, job, and Maya reply variants found in the repository. The shared notification template carries many event types; these are representative states, not separate delivery templates. Dynamic AI wording, arbitrary field combinations, and provider-side Brevo automations are not an exhaustive enumerable catalog. No direct-message chat email producer was found; the conversation preview represents comment replies.

## Review boundary

Only prototype files changed. Existing integrated previews continue to represent application templates. Subjects, preheaders, bodies, CTA labels, and footers here are proposed together. All values and destinations are fictional. Production implementation must preserve permission filtering, actual event facts, recipient timezone, token expiry, and unsubscribe/reply behavior. Maya replies retain literal CONFIRM and CANCEL commands; confirmation proposes one mutation.

## Icons and avatars

PNG icons mark access, calendar changes, conversations, time-sensitive notices, and success receipts. Avatars are 20px with 8px initials and a 4px name gap. Actor avatars appear inline at the start of activity sentences, rather than in a separate top row. The renderer executes packages/lib/src/avatar-color.ts directly, preserving the frontend palette, normalization, and hash. If a person has a valid avatarURL, the renderer uses that image; otherwise it derives initials from the name. A supplied image has initials as alternative text and the name remains readable if images are blocked. The conversation fixture uses a sample portrait from the landing assets, not a real user account. Live avatar URLs still need passing through the production notification payload when this design is integrated.

## Rollout copy locations

Invitation subjects exist in both the event consumer and active invitation worker. Notification copy comes from rules and task handlers; digests come from jobs; Maya content comes from the email agent and deterministic reply processor. Editing template HTML alone will not update every subject or dynamically produced message.

## Preview map

| Preview | Group | Source |
|---|---|---|
| [Sign-in link + code](emails/login-link.html) | Access | apps/server/templates/auth/verification.html |
| [Sign-in link](emails/login-link-only.html) | Access | apps/server/templates/auth/verification.html |
| [Mobile verification code](emails/login-mobile.html) | Access | apps/server/templates/auth/verification_mobile.html |
| [Feedback portal verification](emails/portal-verification.html) | Access | apps/server/templates/feedback/verification.html |
| [Workspace invitation](emails/invitation.html) | Workspace | apps/server/templates/invites/invitation.html |
| [Invitation accepted](emails/invitation-accepted.html) | Workspace | apps/server/templates/invites/acceptance.html |
| [Account inactivity warning](emails/account-inactivity.html) | Workspace | apps/server/templates/users/inactivity_warning.html |
| [Workspace inactivity warning](emails/workspaces-inactivity_warning.html) | Workspace | apps/server/templates/workspaces/inactivity_warning.html |
| [Deletion scheduled · requester](emails/deletion-requested.html) | Workspace | apps/server/templates/workspaces/deletion_scheduled_confirmation.html |
| [Deletion scheduled · other admins](emails/deletion-admin.html) | Workspace | apps/server/templates/workspaces/deletion_scheduled_notification.html |
| [Workspace restored · requester](emails/restored-requester.html) | Workspace | apps/server/templates/workspaces/restored_confirmation.html |
| [Workspace restored · other admins](emails/restored-admin.html) | Workspace | apps/server/templates/workspaces/restored_notification.html |
| [Story assigned](emails/assigned.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Story reassigned](emails/reassigned.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Story unassigned](emails/unassigned.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Added as collaborator](emails/collaborator-added.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Removed as collaborator](emails/collaborator-removed.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Priority changed](emails/priority.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Status changed](emails/status.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Description updated](emails/description.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Due date added](emails/due-date-set.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Due date changed](emails/due-date-changed.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Due date removed](emails/due-date-removed.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Start date changed](emails/start-date.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Sprint changed](emails/sprint.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Estimate changed](emails/estimate.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Collaborators updated](emails/collaborators.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Story renamed](emails/renamed.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Other story updates](emails/story-update.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Story assigned by Maya](emails/maya-assigned.html) | Activity | apps/server/internal/modules/notifications/service/rules.go |
| [Story scheduled](emails/scheduled.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_schedule.go |
| [Scheduled day changed](emails/schedule-day.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_schedule.go |
| [Scheduled time changed](emails/schedule-time.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_schedule.go |
| [Scheduling · needs time](emails/needs-time.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_schedule.go |
| [Scheduling · at risk](emails/at-risk.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_schedule.go |
| [Scheduling · cannot fit](emails/cannot-fit.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_schedule.go |
| [New story comment](emails/comment.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Conversation reply](emails/conversation.html) | Activity | apps/server/internal/modules/notifications/service/rules.go |
| [Mention](emails/mention.html) | Activity | apps/server/internal/modules/notifications/service/rules_story_updates.go |
| [Objective update](emails/objective-update.html) | Strategy | apps/server/internal/modules/notifications/service/rules.go |
| [Key result update](emails/key-result-update.html) | Strategy | apps/server/internal/modules/notifications/service/rules.go |
| [Planning reminder](emails/strategy-planning.html) | Strategy | apps/server/pkg/jobs/strategy_communications.go |
| [Strategy check-in reminder](emails/strategy-check-in.html) | Strategy | apps/server/pkg/jobs/strategy_weekly_check_ins.go |
| [Monthly strategy summary](emails/strategy-monthly.html) | Strategy | apps/server/pkg/jobs/strategy_communications.go |
| [Feedback comment](emails/feedback-comment.html) | Feedback | apps/server/internal/modules/notifications/service/rules.go |
| [Feedback status changed](emails/feedback-status.html) | Feedback | apps/server/internal/modules/notifications/service/rules.go |
| [Feedback update published](emails/feedback-published.html) | Feedback | apps/server/internal/modules/notifications/service/rules.go |
| [Feedback merged](emails/feedback-merged.html) | Feedback | apps/server/internal/modules/notifications/service/rules.go |
| [Grouped activity updates](emails/activity-digest.html) | Digests | apps/server/internal/taskhandlers/notification_handlers.go |
| [Weekly digest · FortyOne](emails/weekly-digest.html) | Digests | apps/server/pkg/jobs/weekly_digest.go |
| [Overdue stories reminder](emails/overdue-stories.html) | Digests | apps/server/pkg/jobs/overdue_stories.go |
| [Overdue objectives reminder](emails/overdue-objectives.html) | Digests | apps/server/pkg/jobs/objective_overdue_email.go |
| [Feedback digest](emails/feedback-digest.html) | Digests | apps/server/pkg/jobs/feedback_digest.go |
| [Weekly note](emails/maya-weekly.html) | Maya | apps/server/pkg/jobs/weekly_digest.go |
| [Answer to a question](emails/maya-answer.html) | Maya | apps/server/internal/modules/emailreply/service/ |
| [Clarification needed](emails/maya-clarify.html) | Maya | apps/server/internal/modules/emailreply/service/ |
| [Proposed change · confirm](emails/maya-confirmation.html) | Maya | apps/server/internal/modules/emailreply/service/ |
| [Change applied](emails/maya-applied.html) | Maya | apps/server/internal/modules/emailreply/service/processor_proposals.go |
| [Change already applied](emails/maya-already-applied.html) | Maya | apps/server/internal/modules/emailreply/service/processor_proposals.go |
| [Change cancelled](emails/maya-cancelled.html) | Maya | apps/server/internal/modules/emailreply/service/processor_proposals.go |
| [Proposal no longer pending](emails/maya-expired.html) | Maya | apps/server/internal/modules/emailreply/service/processor_proposals.go |
| [Item or access changed](emails/maya-conflict.html) | Maya | apps/server/internal/modules/emailreply/service/processor_proposals.go |
| [Earlier preview unavailable](emails/maya-preview-unavailable.html) | Maya | apps/server/internal/modules/emailreply/service/processor.go |
| [Unsupported request](emails/maya-refusal.html) | Maya | apps/server/internal/modules/emailagent/service/prompt.go |
| [Resume proposed change](emails/maya-resumed.html) | Maya | apps/server/internal/modules/emailreply/service/processor_helpers.go |
