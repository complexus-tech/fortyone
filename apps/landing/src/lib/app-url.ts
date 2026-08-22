const DEFAULT_APP_URL = "https://cloud.fortyone.app";

export const APP_URL = (
  process.env.NEXT_PUBLIC_APP_URL ?? DEFAULT_APP_URL
).replace(/\/$/, "");

export const SIGNUP_URL = `${APP_URL}/signup`;

type SignupProvider = "google" | "microsoft";

const getProviderSignupUrl = (provider: SignupProvider) => {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "");
  if (!apiUrl) {
    return SIGNUP_URL;
  }

  try {
    const providerAuthUrl = new URL(`/auth/${provider}`, apiUrl);
    providerAuthUrl.searchParams.set("callbackURL", `${APP_URL}/auth-callback`);
    return providerAuthUrl.toString();
  } catch {
    return SIGNUP_URL;
  }
};

export const GOOGLE_SIGNUP_URL = getProviderSignupUrl("google");
export const MICROSOFT_SIGNUP_URL = getProviderSignupUrl("microsoft");
