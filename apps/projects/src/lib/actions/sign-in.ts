import { getApiUrl } from "@/lib/api-url";

type SignInProvider = "google" | "microsoft";

const signInWithProvider = async (
  provider: SignInProvider,
  callbackUrl = "/auth-callback",
) => {
  if (typeof window === "undefined") {
    throw new Error(`${provider} sign-in is only available in the browser`);
  }

  const apiUrl = getApiUrl();
  const callbackTarget = callbackUrl.startsWith("http")
    ? callbackUrl
    : new URL(callbackUrl, window.location.origin).toString();

  const authUrl = new URL(`/auth/${provider}`, apiUrl);
  authUrl.searchParams.set("callbackURL", callbackTarget);

  window.location.assign(authUrl.toString());
};

export const signInWithGoogle = (callbackUrl = "/auth-callback") =>
  signInWithProvider("google", callbackUrl);

export const signInWithMicrosoft = (callbackUrl = "/auth-callback") =>
  signInWithProvider("microsoft", callbackUrl);
