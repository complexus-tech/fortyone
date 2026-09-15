export const MOBILE_REDIRECT_URI = "fortyone://login";
export const MOBILE_ACCOUNT_DELETION_PATH = "/auth/account-deletion";
const RANDOM_VALUE = /^[A-Za-z0-9_-]{43}$/;

export type MobileAuthRequest = { state: string; codeChallenge: string };

export const parseMobileAuthRequest = (params: {
  state?: string | string[];
  code_challenge?: string | string[];
}): MobileAuthRequest | null => {
  if (
    typeof params.state !== "string" ||
    !RANDOM_VALUE.test(params.state) ||
    typeof params.code_challenge !== "string" ||
    !RANDOM_VALUE.test(params.code_challenge)
  )
    return null;
  return { state: params.state, codeChallenge: params.code_challenge };
};

export const getMobileAuthPath = ({
  state,
  codeChallenge,
}: MobileAuthRequest) =>
  `/auth/mobile?${new URLSearchParams({ state, code_challenge: codeChallenge }).toString()}`;

export const isMobileAuthPath = (value?: string | null) => {
  if (!value?.startsWith("/auth/mobile?")) return false;
  const url = new URL(value, "https://cloud.fortyone.app");
  return (
    url.pathname === "/auth/mobile" &&
    !url.hash &&
    url.searchParams.getAll("state").length === 1 &&
    url.searchParams.getAll("code_challenge").length === 1 &&
    parseMobileAuthRequest({
      state: url.searchParams.get("state") ?? undefined,
      code_challenge: url.searchParams.get("code_challenge") ?? undefined,
    }) !== null
  );
};

const MOBILE_CONTINUATION_PATHS = new Set([
  "/auth-callback",
  "/onboarding/account",
  "/onboarding/create",
  "/onboarding/invite",
  "/onboarding/join",
  "/onboarding/welcome",
]);

// An expired session can wrap the handoff in an onboarding return URL. Keep
// that login email-only even when the outer mobileApp flag was not preserved.
export const isMobileAuthFlow = (callbackUrl?: string | null): boolean => {
  let callback = callbackUrl;
  for (let depth = 0; depth < 5 && callback; depth++) {
    if (
      callback === MOBILE_ACCOUNT_DELETION_PATH ||
      callback === `${MOBILE_ACCOUNT_DELETION_PATH}?mobileApp=true`
    )
      return true;
    if (isMobileAuthPath(callback)) return true;
    if (
      callback.length > 2048 ||
      !callback.startsWith("/") ||
      callback.startsWith("//") ||
      Array.from(callback).some(
        (character) =>
          character === "\\" ||
          character.charCodeAt(0) <= 32 ||
          character.charCodeAt(0) === 127,
      )
    )
      return false;
    const url = new URL(callback, "https://cloud.fortyone.app");
    if (
      (!MOBILE_CONTINUATION_PATHS.has(url.pathname) &&
        !/^\/[a-z0-9][a-z0-9-]*\/settings(?:\/workspace\/members)?$/.test(
          url.pathname,
        )) ||
      url.hash ||
      url.searchParams.getAll("callbackUrl").length !== 1
    )
      return false;
    callback = url.searchParams.get("callbackUrl");
  }
  return false;
};

export const getMobileRedirectURL = (code: string, state: string) => {
  if (!RANDOM_VALUE.test(code) || !RANDOM_VALUE.test(state)) {
    throw new Error(
      "The mobile sign-in response was invalid. Please try again.",
    );
  }
  const url = new URL(MOBILE_REDIRECT_URI);
  url.searchParams.set("code", code);
  url.searchParams.set("state", state);
  return url.toString();
};
