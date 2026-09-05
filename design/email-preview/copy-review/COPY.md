# FortyOne email copy review

65 proposed specimens. All names, dates, metrics, message excerpts, codes, and links are fictional. Production copy remains unchanged.

## Sign-in link + code

- Subject: Sign in to FortyOne
- Preheader: Use the button below to sign in, or enter this code in FortyOne.
- Source: `apps/server/templates/auth/verification.html`

Sign in to FortyOne

Use the button below to sign in, or enter this code in FortyOne.

Your link and code expire in 10 minutes.

482 916

Sign in: https://example.com/fortyone/auth/verify

If you didn’t request this email, you can ignore it.

FortyOne by Complexus LLC


## Sign-in link

- Subject: Sign in to FortyOne
- Preheader: Your sign-in link is ready. It expires in 10 minutes.
- Source: `apps/server/templates/auth/verification.html`

Sign in to FortyOne

Your sign-in link is ready. It expires in 10 minutes.

Sign in: https://example.com/fortyone/auth/verify

If you didn’t request this email, you can ignore it.

FortyOne by Complexus LLC


## Mobile verification code

- Subject: Your sign-in code
- Preheader: Enter this code in FortyOne to sign in.
- Source: `apps/server/templates/auth/verification_mobile.html`

Your sign-in code

Enter this code in FortyOne to sign in.

This code expires in 10 minutes.

482 916

If you didn’t request this email, you can ignore it.

FortyOne by Complexus LLC


## Feedback portal verification

- Subject: Verify your email
- Preheader: Verify your email to continue to Acme’s feedback portal.
- Source: `apps/server/templates/feedback/verification.html`

Verify your email

Verify your email to continue to Acme’s feedback portal.

Use the button below or enter this code. Both expire in 10 minutes.

482 916

Verify email: https://example.com/fortyone/feedback/verify

If you didn’t request this email, you can ignore it.

FortyOne by Complexus LLC


## Workspace invitation

- Subject: Sam invited you to Acme
- Preheader: Join your team on FortyOne. Your invitation expires in 7 days.
- Source: `apps/server/templates/invites/invitation.html`

Your team is ready for you

Sam invited you to join Acme on FortyOne. Join your team to plan projects, track stories, and keep work moving.

Sam Taylor

Join Acme: https://example.com/fortyone/invitations/accept

Your invitation expires in 7 days.

If you weren’t expecting this invitation, you can ignore it.

FortyOne by Complexus LLC


## Invitation accepted

- Subject: Alex has joined Acme
- Preheader: Alex accepted your invitation and is now part of your workspace.
- Source: `apps/server/templates/invites/acceptance.html`

Alex has joined Acme

Alex accepted your invitation and is now part of your workspace.

Alex Morgan

Open Acme: https://example.com/fortyone/workspace

You invited Alex to join Acme.

FortyOne by Complexus LLC


## Account inactivity warning

- Subject: Keep your FortyOne account active
- Preheader: You haven’t signed in for a while. Your account will be deactivated in 30 days if it remains inactive.
- Source: `apps/server/templates/users/inactivity_warning.html`

Keep your FortyOne account active

You haven’t signed in for a while. Your account will be deactivated in 30 days if it remains inactive.

Sign in before then to keep using FortyOne.

Sign in: https://example.com/fortyone/login

This notice concerns your FortyOne account.

FortyOne by Complexus LLC


## Workspace inactivity warning

- Subject: Keep your Acme workspace
- Preheader: Acme has been inactive and is scheduled for deletion in 30 days.
- Source: `apps/server/templates/workspaces/inactivity_warning.html`

Keep your Acme workspace

Acme has been inactive and is scheduled for deletion in 30 days.

Open Acme before then to keep the workspace and its data.

If the workspace stays inactive

It will be deleted, and its data will be permanently removed.

Open Acme: https://example.com/fortyone/workspace

Sent to you as an administrator of Acme.

FortyOne by Complexus LLC


## Deletion scheduled · requester

- Subject: Acme is scheduled for deletion
- Preheader: You requested to delete Acme. Its data will be permanently deleted after the time below.
- Source: `apps/server/templates/workspaces/deletion_scheduled_confirmation.html`

Acme is scheduled for deletion

You requested to delete Acme. Its data will be permanently deleted after the time below.

Changed your mind? Cancel the deletion in workspace settings before then.

Deletion scheduled: 7 September 2026, 14:00 UTC

Manage deletion: https://example.com/fortyone/settings/workspace

Sent to you as an administrator of Acme.

FortyOne by Complexus LLC


## Deletion scheduled · other admins

- Subject: Acme is scheduled for deletion
- Preheader: Sam requested to delete Acme. Its data will be permanently deleted after the time below.
- Source: `apps/server/templates/workspaces/deletion_scheduled_notification.html`

Acme is scheduled for deletion

Sam requested to delete Acme. Its data will be permanently deleted after the time below.

As a workspace administrator, you can cancel the deletion in workspace settings before then.

Deletion scheduled: 7 September 2026, 14:00 UTC

Manage deletion: https://example.com/fortyone/settings/workspace

Sent to you as an administrator of Acme.

FortyOne by Complexus LLC


## Workspace restored · requester

- Subject: Acme has been restored
- Preheader: You restored Acme. The workspace is available again, and its scheduled deletion has been cancelled.
- Source: `apps/server/templates/workspaces/restored_confirmation.html`

Acme has been restored

You restored Acme. The workspace is available again, and its scheduled deletion has been cancelled.

Open Acme: https://example.com/fortyone/workspace

Sent to you as an administrator of Acme.

FortyOne by Complexus LLC


## Workspace restored · other admins

- Subject: Acme has been restored
- Preheader: Sam restored Acme. The workspace is available again, and its scheduled deletion has been cancelled.
- Source: `apps/server/templates/workspaces/restored_notification.html`

Acme has been restored

Sam restored Acme. The workspace is available again, and its scheduled deletion has been cancelled.

Open Acme: https://example.com/fortyone/workspace

Sent to you as an administrator of Acme.

FortyOne by Complexus LLC


## Story assigned

- Subject: Sam assigned you a story
- Preheader: You’re now assigned to this story.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

Sam assigned you a story

Sam Taylor

ACM-142
Improve the onboarding flow
You’re now assigned to this story.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Story reassigned

- Subject: Your story has a new assignee
- Preheader: Sam reassigned this story from you to Jordan.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

Your story has a new assignee

Sam Taylor

ACM-142
Improve the onboarding flow
Sam reassigned this story from you to Jordan.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Story unassigned

- Subject: You’re no longer assigned
- Preheader: Sam removed you as the assignee of this story.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

You’re no longer assigned

Sam Taylor

ACM-142
Improve the onboarding flow
Sam removed you as the assignee of this story.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Added as collaborator

- Subject: You’re now a collaborator
- Preheader: Sam added you as a collaborator on this story.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

You’re now a collaborator

Sam Taylor

ACM-142
Improve the onboarding flow
Sam added you as a collaborator on this story.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Removed as collaborator

- Subject: Your collaboration role changed
- Preheader: Sam removed you as a collaborator on this story.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

Your collaboration role changed

Sam Taylor

ACM-142
Improve the onboarding flow
Sam removed you as a collaborator on this story.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Priority changed

- Subject: A story’s priority changed
- Preheader: Sam changed the priority from Medium to High.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story’s priority changed

Sam Taylor

ACM-142
Improve the onboarding flow
Sam changed the priority from Medium to High.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Status changed

- Subject: A story is ready for review
- Preheader: Sam changed the status from In progress to In review.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story is ready for review

Sam Taylor

ACM-142
Improve the onboarding flow
Sam changed the status from In progress to In review.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Description updated

- Subject: A story’s description changed
- Preheader: Sam updated the description. Open the story to see the latest requirements.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story’s description changed

Sam Taylor

ACM-142
Improve the onboarding flow
Sam updated the description. Open the story to see the latest requirements.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Due date added

- Subject: A story now has a due date
- Preheader: Sam set the due date to 11 September 2026.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story now has a due date

Sam Taylor

ACM-142
Improve the onboarding flow
Sam set the due date to 11 September 2026.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Due date changed

- Subject: A story’s due date changed
- Preheader: Sam moved the due date from 11 September to 14 September 2026.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story’s due date changed

Sam Taylor

ACM-142
Improve the onboarding flow
Sam moved the due date from 11 September to 14 September 2026.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Due date removed

- Subject: A story’s due date was removed
- Preheader: Sam removed the due date. This story no longer has a deadline.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story’s due date was removed

Sam Taylor

ACM-142
Improve the onboarding flow
Sam removed the due date. This story no longer has a deadline.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Start date changed

- Subject: A story’s start date changed
- Preheader: Sam moved the start date from 7 September to 8 September 2026.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story’s start date changed

Sam Taylor

ACM-142
Improve the onboarding flow
Sam moved the start date from 7 September to 8 September 2026.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Sprint changed

- Subject: A story moved to another sprint
- Preheader: Sam moved this story from Sprint 12 to Sprint 13.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story moved to another sprint

Sam Taylor

ACM-142
Improve the onboarding flow
Sam moved this story from Sprint 12 to Sprint 13.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Estimate changed

- Subject: A story’s estimate changed
- Preheader: Sam changed the estimate from 3 to 5 points.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story’s estimate changed

Sam Taylor

ACM-142
Improve the onboarding flow
Sam changed the estimate from 3 to 5 points.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Collaborators updated

- Subject: A story’s collaborators changed
- Preheader: Sam updated the collaborators on this story.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story’s collaborators changed

Sam Taylor

ACM-142
Improve the onboarding flow
Sam updated the collaborators on this story.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Story renamed

- Subject: A story has a new name
- Preheader: Sam renamed “Refresh onboarding” to “Improve the onboarding flow”.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

A story has a new name

Sam Taylor

ACM-142
Improve the onboarding flow
Sam renamed “Refresh onboarding” to “Improve the onboarding flow”.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Other story updates

- Subject: Sam updated a story
- Preheader: Open the story to review the latest changes.
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

Sam updated a story

Sam Taylor

ACM-142
Improve the onboarding flow
Open the story to review the latest changes.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Story assigned by Maya

- Subject: Maya assigned you a story
- Preheader: This story matches your frontend experience and your current availability.
- Source: `apps/server/internal/modules/notifications/service/rules.go`

Maya assigned you a story

ACM-142
Improve the onboarding flow
This story matches your frontend experience and your current availability.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Story scheduled

- Subject: A story has been scheduled
- Preheader: Scheduled for 8 September, 09:00–10:30.
- Source: `apps/server/internal/modules/notifications/service/rules_story_schedule.go`

A story has been scheduled

ACM-142
Improve the onboarding flow
Scheduled for 8 September, 09:00–10:30.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Scheduled day changed

- Subject: A story moved to another day
- Preheader: Moved from 8 September to 9 September, 09:00–10:30.
- Source: `apps/server/internal/modules/notifications/service/rules_story_schedule.go`

A story moved to another day

ACM-142
Improve the onboarding flow
Moved from 8 September to 9 September, 09:00–10:30.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Scheduled time changed

- Subject: A story’s scheduled time changed
- Preheader: Moved from 09:00–10:30 to 11:00–12:30 on 8 September.
- Source: `apps/server/internal/modules/notifications/service/rules_story_schedule.go`

A story’s scheduled time changed

ACM-142
Improve the onboarding flow
Moved from 09:00–10:30 to 11:00–12:30 on 8 September.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Scheduling · needs time

- Subject: A story needs more time
- Preheader: There isn’t enough time allocated to finish this story. Review its estimate and schedule.
- Source: `apps/server/internal/modules/notifications/service/rules_story_schedule.go`

A story needs more time

ACM-142
Improve the onboarding flow
There isn’t enough time allocated to finish this story. Review its estimate and schedule.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Scheduling · at risk

- Subject: A story may miss its deadline
- Preheader: The current schedule puts this story’s deadline at risk. Review the work and available time.
- Source: `apps/server/internal/modules/notifications/service/rules_story_schedule.go`

A story may miss its deadline

ACM-142
Improve the onboarding flow
The current schedule puts this story’s deadline at risk. Review the work and available time.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Scheduling · cannot fit

- Subject: A story couldn’t fit in your schedule
- Preheader: There isn’t enough available time before the deadline. Review the estimate, deadline, or your availability.
- Source: `apps/server/internal/modules/notifications/service/rules_story_schedule.go`

A story couldn’t fit in your schedule

ACM-142
Improve the onboarding flow
There isn’t enough available time before the deadline. Review the estimate, deadline, or your availability.

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## New story comment

- Subject: Sam commented on your story
- Preheader: Sam commented on your story
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

Sam commented on your story

Sam Taylor

ACM-142
Improve the onboarding flow
The updated screens are ready. Could you review the welcome step before we hand this over?

View story: https://example.com/fortyone/stories/ACM-142

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Conversation reply

- Subject: Sam replied to your comment
- Preheader: The updated screens are ready for review.
- Source: `apps/server/internal/modules/notifications/service/rules.go`

Sam replied to your comment

Sam Taylor

Your comment: Can we make the welcome step shorter?

Sam Taylor: Yes, I’ve reduced it to two steps and updated the designs. Could you take another look?

View conversation: https://example.com/fortyone/stories/ACM-142/comments

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Mention

- Subject: Sam mentioned you
- Preheader: Sam mentioned you
- Source: `apps/server/internal/modules/notifications/service/rules_story_updates.go`

Sam mentioned you

Sam Taylor

ACM-142
Improve the onboarding flow
Alex, could you confirm whether these screens cover the sign-in flow?

View mention: https://example.com/fortyone/stories/ACM-142/comments

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Objective update

- Subject: Sam updated an objective
- Preheader: Improve activation is now on track.
- Source: `apps/server/internal/modules/notifications/service/rules.go`

Sam updated an objective

Sam Taylor

Improve activation
Health changed from At risk to On track.

View objective: https://example.com/fortyone/objectives/activation

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Key result update

- Subject: A key result has new progress
- Preheader: First-week activation increased from 42% to 48%.
- Source: `apps/server/internal/modules/notifications/service/rules.go`

A key result has new progress

Sam Taylor

Increase first-week activation to 60%
Progress changed from 42% to 48%.

View key result: https://example.com/fortyone/objectives/activation

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Planning reminder

- Subject: Finish Acme’s strategy for Q4
- Preheader: Your Q4 strategy is missing a few details. Add them to give your team a clear plan for the quarter.
- Source: `apps/server/pkg/jobs/strategy_communications.go`

Finish Acme’s strategy for Q4

Your Q4 strategy is missing a few details. Add them to give your team a clear plan for the quarter.

Still to add: Objectives and key results

Continue planning: https://example.com/fortyone/strategy

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Strategy check-in reminder

- Subject: Two objectives need a check-in
- Preheader: Add a short update so your team can see what’s on track and where support is needed.
- Source: `apps/server/pkg/jobs/strategy_weekly_check_ins.go`

Two objectives need a check-in

Add a short update so your team can see what’s on track and where support is needed.

Improve activation
No check-in in 14 days.

Reduce support response time
At risk · Last checked in 10 days ago.

Review objectives: https://example.com/fortyone/objectives

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Monthly strategy summary

- Subject: Acme’s August strategy review
- Preheader: Here’s how your objectives and delivery progressed in August.
- Source: `apps/server/pkg/jobs/strategy_communications.go`

Acme’s August strategy review

Here’s how your objectives and delivery progressed in August.

Objectives: 3 on track · 1 at risk

Stories completed: 24

Still needs attention: Support response time

Review strategy: https://example.com/fortyone/strategy

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Feedback comment

- Subject: A new reply to your feedback
- Preheader: Thanks for the suggestion. Which filters would you like to save?
- Source: `apps/server/internal/modules/notifications/service/rules.go`

A new reply to your feedback

Sam Taylor

Sam replied to “Add saved views”
Thanks for the suggestion. Which filters would you like to save?

View feedback: https://example.com/fortyone/feedback/saved-views

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Feedback status changed

- Subject: Your feedback is now planned
- Preheader: The status changed from Under review to Planned.
- Source: `apps/server/internal/modules/notifications/service/rules.go`

Your feedback is now planned

Add saved views
The status changed from Under review to Planned.

View feedback: https://example.com/fortyone/feedback/saved-views

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Feedback update published

- Subject: There’s an update on your feedback
- Preheader: Saved views are now available. You can save your filters and return to them from the sidebar.
- Source: `apps/server/internal/modules/notifications/service/rules.go`

There’s an update on your feedback

Add saved views
Saved views are now available. You can save your filters and return to them from the sidebar.

View feedback: https://example.com/fortyone/feedback/saved-views

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Feedback merged

- Subject: Your feedback was merged
- Preheader: Your feedback was merged into “Save and share filtered views”. Open that item to follow future updates.
- Source: `apps/server/internal/modules/notifications/service/rules.go`

Your feedback was merged

Add saved views
Your feedback was merged into “Save and share filtered views”. Open that item to follow future updates.

View feedback: https://example.com/fortyone/feedback/saved-views

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Grouped activity updates

- Subject: Three updates in Acme
- Preheader: Here’s what changed in the work you follow.
- Source: `apps/server/internal/taskhandlers/notification_handlers.go`

Three updates in Acme

Here’s what changed in the work you follow.

ACM-142
Improve the onboarding flow
Sam moved this story to In review.

ACM-156
Add saved views
Jordan assigned this story to you.

ACM-161
Update the billing page
Sam mentioned you in a comment.

View updates: https://example.com/fortyone/notifications

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Weekly digest · FortyOne

- Subject: Your week in Acme
- Preheader: Here’s a quick look at your team’s progress and what needs attention next.
- Source: `apps/server/pkg/jobs/weekly_digest.go`

Your week in Acme

Here’s a quick look at your team’s progress and what needs attention next.

12 stories completed
The onboarding refresh and saved views are ready.

Three stories need attention
Review overdue work before planning next week.

Open Acme: https://example.com/fortyone/workspace

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Overdue stories reminder

- Subject: Two stories are overdue
- Preheader: Review these stories and update their status or due dates.
- Source: `apps/server/pkg/jobs/overdue_stories.go`

Two stories are overdue

Review these stories and update their status or due dates.

ACM-142
Improve the onboarding flow
Due 3 September · In progress

ACM-156
Add saved views
Due 4 September · In review

Review stories: https://example.com/fortyone/my-work

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Overdue objectives reminder

- Subject: An objective needs your attention
- Preheader: The due date has passed. Review progress and update the plan.
- Source: `apps/server/pkg/jobs/objective_overdue_email.go`

An objective needs your attention

The due date has passed. Review progress and update the plan.

Improve activation
Due 31 August · At risk
Key result: Increase first-week activation to 60%. Current progress: 48%.

Review objective: https://example.com/fortyone/objectives/activation

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Feedback digest

- Subject: Updates on your feedback
- Preheader: There’s new activity on two items you follow.
- Source: `apps/server/pkg/jobs/feedback_digest.go`

Updates on your feedback

There’s new activity on two items you follow.

Add saved views
Sam asked which filters you’d like to save.

Export reports as PDF
The status changed to Planned.

View feedback: https://example.com/fortyone/feedback

Manage notification preferences: https://example.com/fortyone/settings/notifications

FortyOne by Complexus LLC


## Weekly note

- Subject: Your weekly note from Maya
- Preheader: Hi Alex,
- Source: `apps/server/pkg/jobs/weekly_digest.go`

Your week in Acme

Hi Alex,

Your team completed 12 stories this week. The onboarding refresh is ready for review, and three stories still need attention.

Ready for review
Sam finished the onboarding screens. Your feedback will help move them into development.

Needs attention
Three stories are overdue. Review what’s blocked before planning next week.

Reply with what you’d like to look into, and I’ll help you work through it.

Maya
Your AI agent at FortyOne

Open Acme: https://example.com/fortyone/workspace

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Answer to a question

- Subject: Here’s where activation stands
- Preheader: First-week activation is at 48%, against a target of 60%.
- Source: `apps/server/internal/modules/emailreply/service/`

Here’s where activation stands

First-week activation is at 48%, against a target of 60%.

The onboarding refresh is in review. The next step is to review Sam’s screens and confirm what’s ready to build.

Maya
Your AI agent at FortyOne

View objective: https://example.com/fortyone/objectives/activation

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Clarification needed

- Subject: Which story should I update?
- Preheader: I found two stories called “Update the welcome email”. Reply with the story ID you mean.
- Source: `apps/server/internal/modules/emailreply/service/`

Which story should I update?

I found two stories called “Update the welcome email”. Reply with the story ID you mean.

ACM-142: Update the welcome email · Product

ACM-208: Update the welcome email · Growth

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Proposed change · confirm

- Subject: Ready for your confirmation
- Preheader: I can change the health of “Improve activation” from At risk to On track.
- Source: `apps/server/internal/modules/emailreply/service/`

Ready for your confirmation

I can change the health of “Improve activation” from At risk to On track.

I haven’t applied this change yet.

Objective: Improve activation

Current health: At risk

Proposed health: On track

Reply CONFIRM to apply this change.
Reply CANCEL to leave it unchanged.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Change applied

- Subject: The objective is now on track
- Preheader: I changed the health of “Improve activation” from At risk to On track.
- Source: `apps/server/internal/modules/emailreply/service/processor_proposals.go`

The objective is now on track

I changed the health of “Improve activation” from At risk to On track.

Reply if you’d like help with another update.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Change already applied

- Subject: That change is already complete
- Preheader: “Improve activation” is already set to On track. No further change was made.
- Source: `apps/server/internal/modules/emailreply/service/processor_proposals.go`

That change is already complete

“Improve activation” is already set to On track. No further change was made.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Change cancelled

- Subject: The proposed change is cancelled
- Preheader: I cancelled the proposed health change for “Improve activation”. I haven’t applied it.
- Source: `apps/server/internal/modules/emailreply/service/processor_proposals.go`

The proposed change is cancelled

I cancelled the proposed health change for “Improve activation”. I haven’t applied it.

Reply if you’d like to try a different update.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Proposal no longer pending

- Subject: Let’s prepare a fresh preview
- Preheader: That proposed change is no longer pending. I haven’t applied it.
- Source: `apps/server/internal/modules/emailreply/service/processor_proposals.go`

Let’s prepare a fresh preview

That proposed change is no longer pending. I haven’t applied it.

Reply with the update you want, and I’ll prepare a new preview.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Item or access changed

- Subject: I couldn’t apply that change
- Preheader: The item or your access changed after the preview, so I couldn’t apply the update.
- Source: `apps/server/internal/modules/emailreply/service/processor_proposals.go`

I couldn’t apply that change

The item or your access changed after the preview, so I couldn’t apply the update.

Reply with what you want to change, and I’ll check it again.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Earlier preview unavailable

- Subject: I need to create a new preview
- Preheader: I couldn’t restore the earlier preview. I haven’t changed anything.
- Source: `apps/server/internal/modules/emailreply/service/processor.go`

I need to create a new preview

I couldn’t restore the earlier preview. I haven’t changed anything.

Reply with the update you want, and I’ll prepare a fresh preview.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Unsupported request

- Subject: I can’t do that by email
- Preheader: I can’t delete a workspace from an email reply. You can manage workspace deletion in FortyOne’s workspace settings.
- Source: `apps/server/internal/modules/emailagent/service/prompt.go`

I can’t do that by email

I can’t delete a workspace from an email reply. You can manage workspace deletion in FortyOne’s workspace settings.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC


## Resume proposed change

- Subject: Your proposed change is ready
- Preheader: Here’s the change we were discussing. I haven’t applied it yet.
- Source: `apps/server/internal/modules/emailreply/service/processor_helpers.go`

Your proposed change is ready

Here’s the change we were discussing. I haven’t applied it yet.

Objective: Improve activation

Change health: At risk → On track

Reply CONFIRM to apply this change.
Reply CANCEL to leave it unchanged.

Maya
Your AI agent at FortyOne

You’re receiving this reply in your conversation with Maya.

FortyOne by Complexus LLC

