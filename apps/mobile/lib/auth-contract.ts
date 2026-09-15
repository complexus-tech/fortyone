export const SESSION_COOKIE_NAME = "fortyone_session";
export const MOBILE_REDIRECT_URI = "fortyone://login";
const RANDOM_VALUE = /^[A-Za-z0-9_-]{43}$/;

export type StoredSession = {
  cookie: string;
  apiOrigin: string;
  expiresAt: number;
  userId: string;
  workspace: string | null;
};
export type SignInTransaction = {
  state: string;
  verifier: string;
  createdAt: number;
};

export const isStoredSession = (value: unknown): value is StoredSession => {
  if (!value || typeof value !== "object") return false;
  const session = value as Partial<StoredSession>;
  return (
    typeof session.cookie === "string" &&
    session.cookie.startsWith(`${SESSION_COOKIE_NAME}=`) &&
    RANDOM_VALUE.test(session.cookie.slice(SESSION_COOKIE_NAME.length + 1)) &&
    typeof session.expiresAt === "number" &&
    Number.isFinite(session.expiresAt) &&
    typeof session.apiOrigin === "string" &&
    /^https?:\/\/[^/?#]+$/.test(session.apiOrigin) &&
    typeof session.userId === "string" &&
    session.userId.length > 0 &&
    (session.workspace === null || typeof session.workspace === "string")
  );
};

// Parse only the one first-party session cookie. Other cookies are never stored
// or replayed, and SecureStore owns expiry and deletion instead of a native jar.
export const parseSessionCookie = (
  header: string | readonly string[] | null,
  apiURL: URL,
  now = Date.now(),
) => {
  const headers = typeof header === "string" ? [header] : header ?? [];
  const sessionCookies = headers
    // iOS may combine Set-Cookie headers, including load-balancer cookies.
    // A separator introduces another name=value; an Expires date's comma does not.
    .flatMap((value) => value.split(/,(?=\s*[^=;,\s]+=)/))
    .map((value) => value.trim())
    .filter((value) => value.split("=", 1)[0].trim() === SESSION_COOKIE_NAME);
  if (sessionCookies.length === 0)
    throw new Error("Sign-in did not return a session cookie.");
  if (sessionCookies.length !== 1)
    throw new Error("Sign-in returned multiple session cookies.");
  const [cookie, ...attributes] = sessionCookies[0]
    .split(";")
    .map((part) => part.trim());
  if (
    !cookie.startsWith(`${SESSION_COOKIE_NAME}=`) ||
    !RANDOM_VALUE.test(cookie.slice(SESSION_COOKIE_NAME.length + 1))
  ) {
    throw new Error("Sign-in returned an invalid session cookie.");
  }
  const values = new Map<string, string>(
    attributes.map((part) => {
      const index = part.indexOf("=");
      return index < 0
        ? [part.toLowerCase(), ""]
        : [part.slice(0, index).toLowerCase(), part.slice(index + 1)];
    }),
  );
  const maxAge = Number(values.get("max-age"));
  const domain = values.get("domain")?.replace(/^\./, "").toLowerCase();
  if (
    !values.has("httponly") ||
    values.get("path") !== "/" ||
    values.get("samesite")?.toLowerCase() !== "lax" ||
    (apiURL.protocol === "https:" && !values.has("secure")) ||
    !Number.isInteger(maxAge) ||
    maxAge <= 0 ||
    maxAge > 31 * 24 * 60 * 60 ||
    (domain &&
      apiURL.hostname !== domain &&
      !apiURL.hostname.endsWith(`.${domain}`))
  ) {
    throw new Error("Sign-in returned an invalid session cookie policy.");
  }
  return { cookie, expiresAt: now + maxAge * 1000 };
};

export const readMobileCallbackCode = (
  callbackURL: string,
  expectedState: string,
) => {
  const url = new URL(callbackURL);
  const code = url.searchParams.get("code");
  if (
    `${url.protocol}//${url.host}${url.pathname}` !== MOBILE_REDIRECT_URI ||
    url.hash ||
    url.username ||
    url.password ||
    !code ||
    !RANDOM_VALUE.test(code) ||
    url.searchParams.getAll("state").length !== 1 ||
    url.searchParams.getAll("code").length !== 1 ||
    url.searchParams.get("state") !== expectedState
  ) {
    throw new Error(
      "The sign-in response does not match this request. Please try again.",
    );
  }
  return code;
};
