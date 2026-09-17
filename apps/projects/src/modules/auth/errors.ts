const SIGN_IN_ERROR_MESSAGES = {
  account_unavailable:
    "We couldn’t sign you in to this account. Contact support for help restoring access.",
  oauth_failed: "We couldn’t complete sign-in. Please try again.",
  oauth_cancelled: "Sign-in was cancelled. Please try again when you’re ready.",
  oauth_expired: "Your sign-in attempt expired. Please start again.",
} as const;

export const getSignInErrorMessage = (error?: string) => {
  if (!error || !Object.hasOwn(SIGN_IN_ERROR_MESSAGES, error)) return undefined;
  return SIGN_IN_ERROR_MESSAGES[error as keyof typeof SIGN_IN_ERROR_MESSAGES];
};
