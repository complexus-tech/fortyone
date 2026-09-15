# iOS release checks

Production signing, builds, TestFlight, and App Store submission belong to the
managed mobile release owner. These checks do not authorize a production build or
upload. Run them against the artifact that will actually be submitted.

## Authentication and account deletion

The mobile browser handoff offers first-party email one-time codes. It preserves
the existing PKCE transaction and validated return destination through login,
signup, invitations, and workspace creation. Google and Microsoft buttons remain
available for ordinary desktop login, but are absent from the mobile entry flow.
Sign in with Apple is not added. Review the final flow against
[guideline 4.8](https://developer.apple.com/app-store/review/guidelines/#login-services)
whenever another mobile login provider is introduced.

Settings exposes permanent account deletion, privacy, terms, and support. The
same account deletion endpoint is available from web account settings:
`DELETE /users/account`. It is authenticated as the current user and is not a
workspace deletion. Confirmation captures the original native session; switching
accounts cannot delete or sign out a replacement session. An accepted deletion
is never automatically retried when device-storage cleanup fails.

Deletion erases the personal profile, private documents, assistant conversations
and memories, personal credentials and memberships. Shared tasks, comments and
reply threads, feedback, activity history, shared documents and referenced files
remain under workspace control with **Former user** attribution. Names or emails
within retained work content and ambiguous older notification text can remain;
the confirmation and public privacy policy disclose this boundary. Active
assignments and leadership references become unassigned. An only administrator
of an active workspace must transfer administration or delete the workspace first.
The API returns an actionable conflict before mutating anything.
Limited security/audit records follow the published retention policy.

The native conflict exposes **Resolve workspace ownership**, which opens
`/auth/account-deletion` on the configured application host. This dedicated page
uses email OTP when authentication is needed, identifies the browser account,
links directly to workspace member and deletion settings, and allows account
deletion after ownership is resolved. Native credentials are never included in
the browser URL. A workspace with a single member can be deleted without inviting
another person. Validate this browser return path using the deployed projects
app as well as the native deletion path.

Calendar event removal, Google grant revocation, email-provider contact deletion,
and unreferenced object removal use background cleanup. An accepted response
distinguishes pending connected-service cleanup from immediate completion; it
does not leave the deleted account usable. Some sealed credentials and the email
needed for contact deletion remain only for the restricted cleanup workflow.

Before shipping this client:

1. Review and apply backend migrations 191, 192 and 193 through the normal deployment
   process, then deploy the API and worker together. Confirm recurring account,
   subscriber, calendar, Google Drive, and attachment cleanup is running. Inspect
   retry age and failures; an unconfigured provider must not be reported as a
   successful remote deletion. Keep account deletion off older API instances
   during rollout because their transaction still removes shared comments and
   history; applying migration 193 alone does not change that behavior.
2. Deploy the projects app's mobile OTP flow, account-deletion resolution page,
   and the updated public privacy policy. These source changes do not change the
   currently hosted pages.
3. Exercise new-account OTP, returning-account OTP, resend/expired codes,
   cancellation, invitations, workspace creation, and app restarts on the signed
   production-configured build. Give App Review a reachable test account and
   instructions for obtaining its code; do not introduce an authentication bypass.
4. Using disposable accounts, verify an only-administrator conflict, ordinary
   deletion, provider cleanup retries, simultaneous session changes, and revoked
   sessions on a second device. Verify intact comment/reply threads, retained
   shared files, Former user attribution, and erased private account content
   against the final API, worker, object storage, and provider accounts.

Shared-content retention remains an App Review item to resolve, not a passing
compliance check. Apple's [account deletion requirements](https://developer.apple.com/support/offering-account-deletion-in-your-app/)
explicitly include shared user-generated content and describe legally required
retention. They do not publish a blanket exception for work applications.
[Atlassian's documented model](https://support.atlassian.com/atlassian-account/docs/delete-your-atlassian-account/)
preserves issues and comments under Former user attribution, but adopting that
model does not establish Apple's acceptance of FortyOne's retention policy.
Review the exact retained data and applicable obligations before submission;
do not promise approval based on another application's behavior.

## Remaining user-content review

Stories and comments are shared user content. The reviewed implementation does
not yet establish a complete moderation workflow. Before claiming compliance
with [guideline 1.2](https://developer.apple.com/app-store/review/guidelines/#user-generated-content),
decide and validate the controls appropriate to private workspaces: content
filtering, durable abuse reporting, timely operator review, blocking abusive
users, and published contact information. A support link or a cosmetic Report
button does not establish those controls. Confirm a real moderation owner and
response process; this is an unresolved release item, not a passing check.

The existing internal-admin account suspension and workspace member-removal
controls can support enforcement. However, story/comment abuse reporting and
user blocking have no existing API to connect to, and comment deletion currently
only permits the author. Product feedback and public-portal contributor blocking
do not cover private workspace comments. The next implementation needs a durable
report queue in the existing admin application, audited moderator actions,
defined block behavior, and server-side content filtering. Identify the reviewer
and response coverage before enabling a reporting flow.

## Native SDK privacy resources

`@sentry/cli` is a direct build-time dependency, pinned to the version required by
the installed Sentry React Native SDK. Its Xcode debug-symbol phase resolves the
CLI from the app before checking `SENTRY_DISABLE_AUTO_UPLOAD`; a transitive-only
dependency fails in pnpm's isolated layout, even for an upload-disabled local
Release build. Keep this version aligned when upgrading the Sentry SDK.

`package.json` sets `expo.autolinking.ios.buildFromSource` to `["expo-image"]`.
Expo SDK 57's autolinker supports this per-module source-build override. It makes
CocoaPods install Expo Image's normal native dependencies and copy the vendor's
privacy resources, rather than using its precompiled SPM dependency frameworks.
It does not disable precompiled React Native or change Android configuration.

The original simulator app contained `SDWebImage.framework` without its privacy
manifest. Apple's [third-party SDK requirements](https://developer.apple.com/support/third-party-SDK-requirements/)
explicitly cover SDWebImage and SDKs that repackage it. An app-level required-reason
declaration does not replace the SDK's own resource.

The locally regenerated CocoaPods graph resolved SDWebImage 5.21.7 within Expo
Image 57.0.5's existing `~> 5.21.0` constraint. Its vendor manifest lives at
`ios/Pods/SDWebImage/WebImage/PrivacyInfo.xcprivacy`; its Core podspec declares
`SDWebImage.bundle`. It declares file timestamps for reason `C617.1`, no tracking,
and no collected data. Both generated Debug and Release resource-copy phases now
include that bundle. The generated `ios/` directory is ignored: do not patch Pods,
copy a hand-written vendor manifest, or depend on a local Podfile edit.

After pulling the source configuration, an existing local native checkout needs
`pod install` in `apps/mobile/ios`. Clean managed builds read the package setting
automatically. Keep the override until the upstream precompiled artifact and a
real built app are verified to contain the vendor resource. Native versions
resolve under Expo's CocoaPods constraints; inspect the generated Podfile.lock
and archive on each dependency update.

## Artifact gate

The verifier requires Python 3 (standard library only). From `apps/mobile`:

```sh
pnpm privacy:verify /absolute/path/FortyOne.app
pnpm privacy:verify /absolute/path/FortyOne.xcarchive
pnpm privacy:verify /absolute/path/device/FortyOne.app --release
```

The first command permits simulator artifacts for local resource verification.
Archives automatically enable the release checks. The command fails if the
app-level manifest is absent, any bundled manifest is malformed, or SDWebImage's
own manifest/resource is missing or lacks its reviewed declaration. Release mode
also checks the bundle identifier, version/build numbers, device SDK 26 or newer,
and removal of Expo's development network-discovery metadata. It never builds,
signs, modifies, uploads, or sends telemetry.

Run the gate after the final archive is created and before managed submission.
It is not automatically invoked by EAS, and passing it is not Apple validation.
Xcode Organizer validation must additionally verify signatures, provisioning,
architecture, entitlements, icons, all required SDK manifests, and required-reason
API coverage. Generate and inspect Xcode's aggregate privacy report.

Apple requires Xcode 26 or later with iOS 26 SDK or later as of April 28, 2026.
Xcode 27 submissions opened September 9, 2026. These are **build SDK** requirements;
the app's current deployment minimum is iOS 16.4. Check Apple's
[current requirements](https://developer.apple.com/news/upcoming-requirements/)
again at submission time, and test the app on its supported OS/device range.

## Privacy inventory and App Store metadata

SDK privacy manifests, approved reasons for restricted APIs, and App Store
Connect's App Privacy answers are separate obligations. Neither an empty
`NSPrivacyCollectedDataTypes` array nor a passing artifact check means that the
FortyOne service collects no data. The release owner must reconcile this working
inventory with the production backend, vendors, retention policy, and actual
enabled features before answering Apple's questionnaire.

| Data flow         | Source behavior                                                                                                                   | Release verification                                                                                                                                                                |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Account/profile   | Authentication and profile updates use first-party API services.                                                                  | Review contact information and identifiers, account linkage, purpose, retention, and deletion.                                                                                      |
| Workspace content | Task titles/descriptions, comments, mentions, and related project content are sent to the API.                                    | Classify user content and identifiers; verify intended sharing and retention.                                                                                                       |
| Local content     | Session credentials use SecureStore; query snapshots use MMKV and drafts use AsyncStorage.                                        | Verify backup behavior, device protection, account isolation, and logout/deletion cleanup. Local-only storage is distinct from server collection.                                   |
| Diagnostics       | Sentry initializes only outside development with a public DSN. JavaScript events are scrubbed; native SDK processing is separate. | Inspect authorized native and JavaScript test payloads, installation identifiers, vendor retention, access, and symbolication. Do not infer native scrubbing from JavaScript tests. |
| Media             | Existing remote images/content can be displayed.                                                                                  | Include actual image hosts and network data flows; do not advertise camera, upload, or photo-library access that is not implemented.                                                |

There is no configured advertising/tracking SDK in the reviewed source. Verify
production vendors' actual use before answering tracking questions; diagnostic
identifiers alone do not imply cross-app advertising tracking. The App Privacy
answers and public privacy policy must accurately describe first-party and
third-party practices. See [Apple's privacy guidance](https://developer.apple.com/app-store/user-privacy-and-data-use/).

## Final managed release evidence

- Confirm the managed Expo project/owner, Apple team, App Store record and bundle
  identifier, distribution certificate, provisioning profile, and unique remote
  build number. Do not substitute a personal Expo/Apple account.
- Confirm production `EXPO_PUBLIC_API_URL` and `EXPO_PUBLIC_APP_URL` point to the
  intended trusted HTTPS services. Public values are embedded in the app; secrets
  must never use that prefix. Check the actual archived bundle, not a development
  `.env.local` file. Missing checked-in EAS values do not prove managed values are
  missing.
- If Sentry is enabled, use managed `SENTRY_ORG`, `SENTRY_PROJECT`, and secret
  `SENTRY_AUTH_TOKEN`; inspect native telemetry and verify source-map/dSYM upload.
  Local and CI builds keep `SENTRY_DISABLE_AUTO_UPLOAD=true`.
- Run the artifact gate, Xcode validation and privacy report; confirm export
  compliance, privacy URL, support URL, age-rating answers, screenshots, and
  reviewer instructions in App Store Connect.
- Validate authentication, account deletion, server errors and offline behavior,
  editors/keyboards, accessibility, and navigation on physical iPhone and iPad.
  The source currently advertises iPad support and iOS 16.4 compatibility.
- Test the signed Release build through the managed review process. Metro exports,
  unit tests, a simulator launch, and the resource-only native check are not a
  signed archive, TestFlight, or App Store review result.
