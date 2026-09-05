# Email consolidation and timing proposal

Reviewed 5 September 2026. Proposal only: no delivery behavior, preferences, schedules, or production templates changed. Findings come from the current repository and its installed scheduler dependency, not production delivery logs or Joseph’s inbox. Provider-managed Brevo campaigns and automations are outside this source audit.

## Recommendation

Keep essential transactions and Maya’s direct replies immediate. Consolidate routine workspace communication into a morning briefing, with a limited one-hour catch-up for direct collaboration. Weekly planning, deadline reminders, and strategy updates should contribute sections to those deliveries instead of independently sending overlapping emails.

The 65 design previews are examples of content and states, not 65 independently scheduled email jobs. In particular, the application’s weekly digest already sends as Maya; the gallery’s FortyOne and Maya weekly examples do not establish two weekly production jobs.

## Current behavior

| Family | Current trigger and audience | Sender / overlap |
|---|---|---|
| Login links, mobile codes, portal verification | On request | FortyOne; preserve immediate delivery and expiry. |
| Invitations and acceptance receipts | On the respective event | FortyOne; acceptance is a candidate for batching, but the invitation itself is actionable. |
| Workspace deletion/restoration receipts and admin notices | On the respective event | FortyOne; preserve separate factual lifecycle notices. |
| Account/workspace inactivity warnings | Sunday 02:00 / 01:00 UTC jobs, when eligible | FortyOne; keep distinct from optional work summaries. |
| Activity notification digest | One unique delayed task per recipient/workspace, one hour after the first pending event | Usually FortyOne. Contains only currently eligible unread, unsent notifications. No local working-hours or general daily frequency cap was found in this path. |
| Weekly workspace digest | Monday 08:00 UTC; eligible active admins/members; meaningful signals required | Maya. Includes unread mentions/replies, overdue stories, upcoming stories, and objective deadline signals. |
| Story deadline reminders | Daily 09:00 UTC, including weekends; assigned stories | Maya. Selected reminders at three days before, one day before, due day, and up to three days overdue can repeatedly describe the same story. |
| Objective/key-result deadline reminders | Monday 10:00 UTC; eligible objective leads | Maya. Overlaps the weekly digest and strategy check-ins. |
| Weekly strategy check-in | Generated Wednesday during the recipient’s local 09:00 hour; owners/leads with stale or at-risk items | Goes through the hour-delayed notification digest, so typically arrives around 10:00 local or in an already queued batch. Strategy content makes that digest send as Maya. Staleness threshold is seven days. |
| Strategy planning reminder | Local 09:00 hour, seven days before the next quarter, if required strategy foundations are missing; admins | Uses the same delayed digest route. |
| Monthly strategy summary | First of the month during local 09:00 hour, when there is a signal; admins | Uses the same delayed digest route. |
| Feedback reviewer digest | Daily or weekly subscription; due from 09:00 local, weekly period begins Monday; hourly polling | Maya. New incoming portal/widget/integration submissions for internal board reviewers; has delivery claims and subscription cursors. This is different from customer-facing feedback replies/status notifications. |
| Maya answers, clarifications, proposals, and confirmation/cancellation receipts | In response to an email conversation | Maya. Keep these in their existing conversation, with explicit confirmation semantics. |

The fixed scheduler uses UTC (`asynq.NewScheduler(redisOpt, nil)` and the dependency’s UTC default). For an eligible Harare recipient, Monday weekly, story, and objective jobs can therefore deliver at roughly 10:00, 11:00, and 12:00 respectively. These are job start times, not guaranteed inbox arrival times.

## Why information repeats

1. **Each family decides independently.** The notification digest records `email_sent_at`; deadline jobs compute their own snapshots and send directly. There is no shared content-level suppression or recipient frequency budget across these paths.
2. **Unread is not the same as not previously emailed.** The weekly aggregate checks `read_at`, but not `email_sent_at`. An already emailed mention can still contribute to the weekly summary if it remains unread in the app.
3. **Maya is used across several senders’ purposes.** Weekly, story, objective, feedback, and strategy guidance all use Maya. A strategy item inside an ordinary notification batch can change the batch’s sender and primary CTA to Maya/strategy.
4. **Recurring eligibility is not a material change.** A still-overdue story or stale objective can qualify again without new information. Strategy records are email-only, so the normal in-app unread check is not a sufficient acknowledgement mechanism for those notices.
5. **Preferences mostly toggle categories on/off.** All notification email categories default to enabled. They do not express a shared routine-email frequency. Feedback reviewers have separate daily/weekly/off subscription settings.

The one-hour mechanism is a good batching foundation. It is a window beginning with the first queued event, not a guarantee that every event waits an hour: an event arriving near the end of the window can be included immediately at that window’s send. Keep that fixed window rather than continually resetting it and postponing delivery.

## Proposed default delivery policy

Times below are starting hypotheses for this product, in each recipient’s local timezone.

| Delivery | Content | Proposed schedule |
|---|---|---|
| Essential messages | Authentication, invitations, deletion/restoration notices; Maya’s requested replies and action receipts | Immediate. Outside routine budgets. Inactivity warnings remain separate; queue eligible warnings for a local daytime slot without shortening the promised notice period. |
| Morning briefing from Maya | Unread routine changes, material deadline milestones, invitation acceptance, and applicable reviewer/strategy sections | 09:00 on working days, only when useful content exists. One per workspace, under a shared recipient budget. |
| Direct activity catch-up from FortyOne | Unread mentions, replies, and new assignments that need the recipient’s attention | Keep the one-hour window during working hours. If a briefing is due sooner, include them there. Suppress already read or otherwise acknowledged events before sending. |
| Monday edition of the briefing | This week’s priorities, relevant due work, objective/key-result check-ins, and a brief previous-week summary where supported by facts | Monday 09:00; replaces the standalone weekly digest and Monday objective email. It also absorbs that morning’s story reminders. |
| Strategy follow-up section | A new risk, meaningful escalation, or a still-unanswered request eligible for another reminder | In the next briefing. Stop the automatic extra Wednesday email for unchanged Monday content. |
| Monthly/planning sections | Previous-month strategy review, or missing next-quarter foundations | Admins only, inside the first-working-day briefing / the appropriate briefing before the planning deadline. No second same-day summary. |
| Feedback reviewer section | New submissions awaiting review | Honor the existing daily/weekly/off board subscriptions; merge into the corresponding briefing for that same recipient and workspace. Keep external contributors’ replies/status notices in their own audience and delivery flow. |

A balanced default would allow **at most two routine emails per recipient per local working day**, counting briefings and direct catch-ups together. Pending content stays queued for the next eligible delivery; it is never marked delivered merely because the budget was exhausted. Essential notices and responses to the user’s own Maya conversation do not consume that budget. Users who need more frequent collaboration email can opt into a higher frequency; daily-only and off remain available.

Use an initial routine-delivery window of 09:00–18:00 Monday–Friday, then honor explicitly configured working days/hours. Overnight activity joins the next morning’s briefing. Use the user’s saved IANA timezone and daylight-saving rules; use a documented fallback when it is missing. Keep message content and reply authorization scoped to one workspace; do not combine unrelated workspaces into a single Maya reply thread. A recipient-level budget should fairly schedule competing workspaces, and the settings should make delayed catch-ups clear.

## Consolidation rules

- Group multiple changes to the same story into one compact row showing the latest state. Preserve separate actionable mentions/replies rather than losing them inside a generic “updated” label.
- Select and deduplicate facts before asking AI to write copy. Different wording must not make an old fact eligible again.
- Record which notification IDs, entity versions, deadline milestones, and check-in requests each successful delivery covered. Use that record across every routine sender.
- Suppress repeated details on the same day. For unchanged risk/check-in requests, start with a seven-day repeat interval; allow a genuinely new deadline milestone or escalation sooner. A weekly aggregate can summarize outstanding work without replaying every already delivered event.
- Recheck current permissions, email preferences, read/acknowledgement state, item completion, and the latest item values before composing the final email. A resolved issue must not survive solely because it was captured in an earlier strategy snapshot.
- On Monday, the weekly briefing wins over separate routine summaries. On the first working day of a month, add an admin strategy section to that delivery. A near-due direct catch-up should merge into the morning send rather than land immediately afterwards.
- Show at most five actionable rows, with a count and link for the rest. An empty or all-suppressed briefing sends nothing. Long reviewer queues can link to the full board.
- A successful email is not proof that the user read it. Keep “delivered,” “read,” and “acknowledged/resolved” separate.

## Broader copy/content changes

Make the weekly email a short planning note: priorities for the coming week, one or two concrete decisions, and the relevant context. The current weekly source has counts rather than the rich named-story examples in the gallery, so implementation needs trusted scoped item data before promising specific names or completion highlights. The current `objective_risks` count is deadline-based; it should not be described as a health assessment unless actual health data is added.

Make strategy check-ins specific: which objective needs input, the last check-in, what changed, and the requested action. Remove repeated generic prompts to “review strategy.” Distinguish an unchanged stale check-in from a newly at-risk objective.

Correct the feedback digest specimen during implementation: the current reviewer job summarizes new inbound submissions, whereas the gallery’s general “updates on your feedback” example describes contributor updates. Preserve those distinct audiences even if they share visual components.

## Implementation sequence after agreement

1. **Measure the current baseline.** Inspect a representative 2–4 weeks of provider/delivery history by recipient, workspace, source, time, and message ID. Compare repeated entity facts and user actions; check Brevo-managed automations separately. Source inspection explains possible overlap but does not prove which exact messages Joseph received.
2. **Create one routine-email decision layer.** Reuse the existing notification batching, feedback delivery claims/cursors, and permission-scoped queries. Add shared delivery coverage and per-recipient budgets. Existing jobs should produce candidate content, not independently call the mailer once their replacement is enabled.
3. **Consolidate Monday first.** Replace weekly + story + objective sends with one briefing for an enabled cohort. Gate old and new paths together so rollout itself cannot double-send.
4. **Bring strategy and feedback into the briefing.** Preserve audience, subscription frequency, and Maya reply scope. Add local-time scheduling and explicit delivery preferences without re-enabling anything users disabled.
5. **Apply approved templates to the final families.** Render realistic combined examples, including long names, missing photos, many changes to one story, and empty sections. Preserve immediate transaction and confirmation copy.
6. **Validate and gradually release.** Test simultaneous jobs, read-before-send, resolved-before-send, retries, timezone/DST boundaries, quiet hours, monthly/Monday collisions, unsubscribe changes, and multi-workspace budgets. Use durable claims and stable message identities; do not assume an SMTP Message-ID alone guarantees deduplication after an ambiguous send failure.

## Choosing and measuring send times

09:00 is a product-fit starting point, not a proven universal best time. Mailchimp’s research discusses a broad marketing-email peak around 10:00 in recipients’ timezones and its optimization tool uses audience engagement data. That is useful context, not evidence that FortyOne’s work notifications should wait for a marketing optimum. See [Mailchimp’s timing research](https://mailchimp.com/resources/insights-from-mailchimps-send-time-optimization-system/) and [how its optimization works](https://mailchimp.com/help/use-send-time-optimization/).

After fixing duplication, compare 09:00 versus 10:00 for the briefing with stable recipient-level cohorts and enough data to make a useful comparison. Evaluate emails per active user, repeated-item rate, notification disable/unsubscribe rate, useful clicks, replies, and completed in-app actions. Track time to notice direct collaboration as a guardrail. Do not choose the winner from open rate alone, or optimize invitation/login delivery for campaign engagement.

## Source map

- `apps/server/internal/bootstrap/worker/bootstrap.go:147` and `scheduler.go:226`: scheduler setup and recurring schedules.
- `apps/server/pkg/tasks/notifications.go:23` and `:80`: one-hour coalescing and unique tasks.
- `apps/server/internal/modules/notifications/service/create.go`: active notification producer uses the digest queue; the legacy single-email enqueue function has no current non-test callers found.
- `apps/server/internal/modules/notifications/repository/queries/delivery.sql:232`: unread/unsent delivery filtering.
- `apps/server/internal/modules/notifications/repository/queries/weekly_digest.sql:79`: weekly unread signals and scoped aggregates.
- `apps/server/internal/modules/notifications/domain/preferences.go`: defaults and email-only categories.
- `apps/server/pkg/jobs/weekly_digest.go:37` and `:247`: eligibility and Maya sender.
- `apps/server/internal/modules/stories/repository/queries/guidance.sql` and `objectives/repository/queries/guidance.sql`: deadline audiences and repeated milestone eligibility.
- `apps/server/pkg/jobs/strategy_weekly_check_ins.go:78` and `strategy_communications.go`: local scheduling, planning, and monthly strategy.
- `apps/server/internal/taskhandlers/notification_digest_copy.go:177`: strategy content changes the digest sender to Maya.
- `apps/server/pkg/jobs/feedback_digest.go:212` and `internal/modules/feedback/repository/queries/digest.sql`: reviewer scheduling, new-submission scope, delivery claims, and cursors.
- `apps/server/internal/eventconsumer/email_events.go`, `workspace_notifications.go`, `internal/bootstrap/worker/invitation_email.go`, and `internal/modules/emailreply/service/`: immediate transactional and conversational paths.
