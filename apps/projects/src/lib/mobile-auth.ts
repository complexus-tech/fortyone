export const MOBILE_REDIRECT_URI = "fortyone://login";
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
