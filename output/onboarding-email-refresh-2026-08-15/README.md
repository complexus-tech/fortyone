# FortyOne onboarding email refresh

This package replaces the current feature-tour drip with an activation journey and a distinct FortyOne point of view. It includes paste-ready Brevo HTML, primary and challenger subject lines, current product screenshots, and an interactive preview.

Open `preview.html` to review every template. The HTML files are in `templates/`, and email-safe JPEG exports of the product screenshots are in `assets/`.

## Creative territory

> **Most project tools start with a Task. FortyOne starts with the reason it deserves one.**

The category is crowded with “AI teammate,” “all-in-one,” “work smarter,” “focus on what matters,” and “align and ship faster.” This sequence deliberately avoids that language.

The FortyOne territory is the decision trail:

```text
customer signal → decision → Objective → Task → owner → real week
```

The voice is calm contrarian: one recognizable tension, one concrete consequence, one product mechanism, and one action. It should feel intelligent and compressed—not loud, breathless, or artificially quirky.

Three recurring ideas make the campaign recognizable:

- Every Task should know why it exists.
- AI should show its work before changing yours.
- A plan is not realistic until it fits somebody’s week.

## The change that matters most

Do not start the workspace-owner series from the existing account-created Brevo list alone.

At that point, FortyOne does not yet know whether the person will create a workspace, accept an invitation, or only use the public feedback experience. Email signups can also reach that list before `NAME` exists. The current five-minute email therefore promises a workspace experience to recipients who may not have one.

Use two entry points instead:

1. **Account setup rescue** — only when the backend confirms `workspace_count = 0`, there is no pending invitation, and the account is not feedback-only. Send `00-finish-workspace-setup.html` after 30–60 minutes.
2. **Workspace activation** — start when a workspace is created. The existing workspace-created Brevo task already supplies `NAME`, `WORKSPACE_NAME`, and `WORKSPACE_SLUG`, which makes accurate copy and deep links possible.

The supplied rescue template is not safe to deploy from Brevo list 6 without those conditions.

## Product onboarding friction the emails cannot fix

The repository audit found several first-run problems that can suppress activation even after the email copy improves:

- A normal direct signup lands on the feedback-oriented `/account` screen. Workspace creation is presented as an optional card rather than the next onboarding step.
- The “receive product updates” checkbox on the profile step is local UI state and is not submitted.
- The five tutorial Tasks still use some older “Story” language, while the product now defaults to Task.
- The walkthrough asks someone to create their “first Task” even though provisioning already assigned five tutorial Tasks.
- The welcome screen’s Build your team, Integrations, and Shortcuts cards all lead to My Work instead of their named destinations.

Treat these as a companion activation backlog. Better emails should not be used to paper over a first session that sends people to the wrong place.

## Recommended journey

| Moment | Condition | Send | Primary action |
| --- | --- | --- | --- |
| 5 minutes after workspace creation | No user-created Task | 01 — First real Task | Add the Task someone will ask about |
| 12–24 hours after workspace creation | Workspace creator; frequency cap permits | 02 — Founder question | Reply to Joseph |
| After first real Task, or day 2 | Task exists and Maya has not been used | 03 — Maya planning | Give Maya the unfinished plan |
| Day 3–5 | Real work exists; no Objective | 04 — Objective | Give the Task a reason |
| Day 4–6 | Admin; one active member; no pending invites | 05 — Invite | Make one useful handoff |
| Day 4–6 alternate | More than one active member; calendar not connected | 05B — Calendar | Plan against the real week |
| Around day 7 | Still inactive | Send one human-help note, then stop | Reply for help |

Apply a frequency cap of one educational onboarding email per recipient per local day. Suppress each email as soon as its target action is complete. Do not number the emails or promise what arrives next because recipients will take different branches.

### Founder timing

The current two-hour delay makes the founder note feel automated because it follows the welcome immediately. Start by testing the following morning (12–24 hours). If you retain the two-hour timing, keep the note plain-text-like, send it from Joseph, and make the Reply-To `joseph@fortyone.app`.

## Activation definition

Use **first user-created Task** as the version-one email milestone, not `task_count > 0`: new workspaces are provisioned with five tutorial Tasks.

Recommended candidate event:

```text
first_real_task_created
  workspace_id
  user_id
  source = web | maya | api | integration
  occurred_at
```

This is a practical starting signal, not a proven north star. Validate that it predicts D7/D14/D30 retained workspace activity, then refine the definition. A stronger later definition may combine a real Task with an ownership or collaboration action.

## Brevo data required for branching

The current integration mostly adds a contact to lists. Add event or contact-state updates for:

- `WORKSPACE_SLUG`
- `WORKSPACE_ROLE`
- `REAL_TASK_COUNT`
- `MAYA_USED`
- `OBJECTIVE_COUNT`
- `MEMBER_COUNT`
- `PENDING_INVITATION_COUNT`
- `CALENDAR_CONNECTED`
- `LAST_ACTIVE_AT`

Because one Brevo contact currently stores only one `WORKSPACE_SLUG`, a user who creates multiple workspaces can overwrite the previous value. For durable multi-workspace automation, send workspace-scoped event payloads and build links from event data rather than mutable contact attributes.

## Deep links

The templates use exact workspace destinations with deliberate campaign parameters:

- My Work: `https://{{ contact.WORKSPACE_SLUG }}.fortyone.app/my-work`
- Maya: `https://{{ contact.WORKSPACE_SLUG }}.fortyone.app/maya`
- Roadmap: `https://{{ contact.WORKSPACE_SLUG }}.fortyone.app/roadmap`
- Members: `https://{{ contact.WORKSPACE_SLUG }}.fortyone.app/settings/workspace/members`
- Calendar connection: `https://{{ contact.WORKSPACE_SLUG }}.fortyone.app/settings/account/calendar`

Do not send action CTAs to the generic `https://www.fortyone.app/login` redirect. Remove the existing `utm_source=chatgpt.com` parameter from documentation links.

## Product-truth guardrails applied to the copy

- The default term is **Task**, not Story.
- The first action is “add your first real Task” because tutorial Tasks already exist.
- Maya prepares answers and proposed changes; the recipient reviews important changes before they are applied.
- The sequence does not claim that Maya can create or plan Sprints.
- The sequence does not claim that Task completion automatically updates Key Result values.
- Custom terminology is not a first-value action and is not retained on every paid plan, so it is removed from onboarding.
- Voice is real but limited and is not used as the lead activation promise.
- Current positioning connects strategy and customer feedback to work a team can deliver; it is not reduced to generic “AI project management.”

## Design system

- 600px email shell, warm off-white page, white content surface
- Restrained FortyOne orange as an accent, not a decorative gradient
- System-safe typography and readable 15px body copy
- One job and one CTA per product email
- At most one real product screenshot, directly supporting the requested action
- A deliberately plain founder email with no banner, screenshot, or button
- Live text for all essential content, descriptive image alt text, and a visible unsubscribe link

The templates currently reference the public FortyOne WebP screenshots so they can be pasted and previewed immediately. For maximum Outlook compatibility, upload the supplied JPEGs from `assets/` to Brevo and replace those image URLs with the Brevo-hosted versions.

## Measurement plan

Primary:

- workspace activation within 24 and 48 hours
- median time to first real Task
- D7, D14, and D30 retained workspace activity

Supporting:

- deep-link click-through rate
- target action completed within 30 minutes and 24 hours of a click
- founder-email reply rate
- unsubscribe, complaint, bounce, and delivery-failure rates

Treat opens as directional because privacy features can preload tracking pixels. If volume allows, use a holdout and compare product activation rather than only subject-line opens.

## Initial experiment

1. Hold out 10% of eligible new workspace creators from non-transactional onboarding.
2. Split the remaining audience between the screenshot version and a no-screenshot version.
3. Keep subject, sender, timing, and CTA destination identical within each test.
4. Choose the winner on incremental first-real-Task completion, not click-through rate alone.

Useful source guidance: [Amplitude on activation](https://amplitude.com/blog/getting-started-activation), [Customer.io on exit conditions](https://docs.customer.io/journeys/campaign-exit-conditions/), [Intercom on one-action onboarding messages](https://www.intercom.com/blog/designing-onboarding-message-schedule/), [Mailchimp on Apple Mail Privacy Protection](https://mailchimp.com/help/apple-privacy-faq/), and [Gmail sender requirements](https://support.google.com/mail/answer/81126?hl=en).

## Repository evidence

- Account and workspace contacts enter different Brevo lists: `apps/server/internal/taskhandlers/onboarding.go`
- New-account redirect decisions: `apps/projects/src/utils/index.ts`
- Feedback-oriented account screen: `apps/projects/src/modules/public-portal/account-page.tsx`
- Workspace creation, default team, and trial enrollment: `apps/server/internal/modules/workspaces/service/workspaces.go`
- Five provisioned tutorial Tasks: `apps/server/internal/modules/workspaces/service/seed.go`
- Default Task terminology: `apps/server/internal/modules/workspaces/repository/commands.go`
- Current email design tokens: `apps/server/pkg/mailer/styles.go`
- Current positioning and capability claims: `apps/landing/src/modules/home/`

## Template inventory

- `00-finish-workspace-setup.html` — conditional account-only rescue
- `01-first-real-task.html` — workspace activation
- `02-founder-question.html` — personal founder note
- `03-maya-planning.html` — use Maya with real work
- `04-connect-work-to-objective.html` — connect execution to an outcome
- `05-invite-one-teammate.html` — collaboration branch
- `05b-connect-calendar.html` — alternate branch for an already collaborative workspace

Each file begins with comments containing its sender, subject, preheader, audience condition, and CTA.
