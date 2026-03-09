import { getApiUrl } from "@/lib/api-url";

export const signInWithGoogle = async (callbackUrl = "/auth-callback") => {
  if (typeof window === "undefined") {
    throw new Error("Google sign-in is only available in the browser");
  }

  const apiUrl = getApiUrl();
  const callbackTarget = callbackUrl.startsWith("http")
    ? callbackUrl
    : new URL(callbackUrl, window.location.origin).toString();

  const authUrl = new URL("/auth/google", apiUrl);
  authUrl.searchParams.set("callbackURL", callbackTarget);

  window.location.assign(authUrl.toString());
};
