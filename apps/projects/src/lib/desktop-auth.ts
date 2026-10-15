export const DESKTOP_REDIRECT_URI = "fortyone-desktop://login";
export const DESKTOP_DEV_REDIRECT_URI = "fortyone-desktop-dev://login";
const RANDOM_VALUE = /^[A-Za-z0-9_-]{43}$/;

export type DesktopAuthRequest = {
  state: string;
  codeChallenge: string;
  redirectUri?: string;
};

const isDesktopRedirectUri = (value: string) =>
  value === DESKTOP_REDIRECT_URI ||
  (process.env.NODE_ENV === "development" &&
    value === DESKTOP_DEV_REDIRECT_URI);

export const parseDesktopAuthRequest = (params: {
  state?: string | string[];
  code_challenge?: string | string[];
  redirect_uri?: string | string[];
}): DesktopAuthRequest | null => {
  if (
    typeof params.state !== "string" ||
    !RANDOM_VALUE.test(params.state) ||
    typeof params.code_challenge !== "string" ||
    !RANDOM_VALUE.test(params.code_challenge) ||
    (params.redirect_uri !== undefined &&
      (typeof params.redirect_uri !== "string" ||
        !isDesktopRedirectUri(params.redirect_uri)))
  ) {
    return null;
  }
  return {
    state: params.state,
    codeChallenge: params.code_challenge,
    ...(params.redirect_uri ? { redirectUri: params.redirect_uri } : {}),
  };
};

export const getDesktopAuthPath = ({
  state,
  codeChallenge,
  redirectUri,
}: DesktopAuthRequest) => {
  const params = new URLSearchParams({ state, code_challenge: codeChallenge });
  if (redirectUri) params.set("redirect_uri", redirectUri);
  return `/auth/desktop?${params.toString()}`;
};

export const isDesktopAuthPath = (value?: string | null) => {
  if (!value?.startsWith("/auth/desktop?") || value.length > 256) return false;
  if (
    Array.from(value).some(
      (character) =>
        character === "\\" ||
        character.charCodeAt(0) <= 32 ||
        character.charCodeAt(0) === 127,
    )
  )
    return false;
  const url = new URL(value, "https://cloud.fortyone.app");
  return (
    url.pathname === "/auth/desktop" &&
    !value.includes("#") &&
    Array.from(url.searchParams.keys()).every((key) =>
      ["state", "code_challenge", "redirect_uri"].includes(key),
    ) &&
    url.searchParams.getAll("state").length === 1 &&
    url.searchParams.getAll("code_challenge").length === 1 &&
    url.searchParams.getAll("redirect_uri").length <= 1 &&
    parseDesktopAuthRequest({
      state: url.searchParams.get("state") ?? undefined,
      code_challenge: url.searchParams.get("code_challenge") ?? undefined,
      redirect_uri: url.searchParams.get("redirect_uri") ?? undefined,
    }) !== null
  );
};

export const getDesktopRedirectURL = (
  code: string,
  state: string,
  redirectUri = DESKTOP_REDIRECT_URI,
) => {
  if (
    !RANDOM_VALUE.test(code) ||
    !RANDOM_VALUE.test(state) ||
    !isDesktopRedirectUri(redirectUri)
  ) {
    throw new Error(
      "The desktop sign-in response was invalid. Please try again.",
    );
  }
  const url = new URL(redirectUri);
  url.searchParams.set("code", code);
  url.searchParams.set("state", state);
  return url.toString();
};
