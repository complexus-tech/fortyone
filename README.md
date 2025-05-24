# Projects API

## Running the project locally

## MailerLite Integration

This service integrates with MailerLite to manage subscriber onboarding. When a user registers, they are automatically added to a MailerLite onboarding group.

### Required Environment Variables

Add these environment variables to your `.env` file:

```bash
# MailerLite Configuration
APP_MAILERLITE_API_KEY=your_mailerlite_api_key_here
APP_MAILERLITE_ONBOARDING_GROUP_ID=your_onboarding_group_id_here
```

### Getting MailerLite Credentials

1. **API Key**: Get your API key from MailerLite Dashboard → Settings → Developer API
2. **Group ID**: Create an "Onboarding" group in MailerLite and note the group ID

### How It Works

1. When a user registers via Google Auth, the `EnqueueUserOnboardingStart` task is created
2. The worker processes this task after a 30-minute delay
3. The onboarding handler uses the MailerLite service to:
   - Create/update the subscriber in MailerLite using their email and full name
   - Add the subscriber to the configured onboarding group
4. If MailerLite integration fails, it's logged but doesn't fail the entire onboarding process

### Architecture

The MailerLite integration follows a clean separation of concerns:

- **MailerLite Service** (`pkg/mailerlite`): Provides generic methods for subscriber management and group assignment
- **Onboarding Handler** (`internal/taskhandlers/onboarding.go`): Contains business logic for adding users to the onboarding group
- **Worker Configuration**: Manages API credentials and group IDs through environment variables
