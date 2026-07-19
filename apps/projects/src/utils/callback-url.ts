const FORTYONE_DOMAIN = "fortyone.app";
const DEFAULT_AUTH_HOST = "cloud.fortyone.app";
const MAX_CALLBACK_URL_LENGTH = 2048;

const hasUnsafeCallbackCharacters = (value: string) =>
  Array.from(value).some((character) => {
    const characterCode = character.charCodeAt(0);
    return character === "\\" || characterCode <= 31 || characterCode === 127;
  });

const isLocalHostname = (hostname: string) =>
  hostname === "localhost" ||
  hostname === "127.0.0.1" ||
  hostname.endsWith(".localhost");

const isFortyOneHostname = (hostname: string) =>
  hostname === FORTYONE_DOMAIN || hostname.endsWith(`.${FORTYONE_DOMAIN}`);

export const getSafeCallbackUrl = (callbackUrl?: string | null) => {
  const value = callbackUrl?.trim();

  if (!value) return undefined;
  if (value.length > MAX_CALLBACK_URL_LENGTH) return undefined;
  if (hasUnsafeCallbackCharacters(value)) return undefined;
  if (value.startsWith("/") && !value.startsWith("//")) return value;

  try {
    const parsed = new URL(value);
    const hasSafeProtocol =
      parsed.protocol === "https:" || parsed.protocol === "http:";

    if (!hasSafeProtocol) return undefined;
    if (parsed.username || parsed.password) return undefined;
    if (parsed.protocol === "http:" && !isLocalHostname(parsed.hostname)) {
      return undefined;
    }
    if (isFortyOneHostname(parsed.hostname)) return parsed.toString();
    if (
      process.env.NEXT_PUBLIC_DOMAIN !== FORTYONE_DOMAIN &&
      isLocalHostname(parsed.hostname)
    ) {
      return parsed.toString();
    }
  } catch {
    return undefined;
  }

  return undefined;
};

export const withCallbackUrl = (path: string, callbackUrl?: string | null) => {
  const safeCallbackUrl = getSafeCallbackUrl(callbackUrl);

  if (!safeCallbackUrl) return path;

  const separator = path.includes("?") ? "&" : "?";
  return `${path}${separator}callbackUrl=${encodeURIComponent(safeCallbackUrl)}`;
};

export const getAuthCallbackPath = (callbackUrl?: string | null) =>
  withCallbackUrl("/auth-callback", callbackUrl);

export const getLoginUrl = (callbackUrl?: string | null) => {
  const loginPath = withCallbackUrl("/", callbackUrl);

  if (process.env.NEXT_PUBLIC_DOMAIN !== FORTYONE_DOMAIN) {
    return loginPath;
  }

  const authHost = process.env.NEXT_PUBLIC_AUTH_HOST ?? DEFAULT_AUTH_HOST;
  return `https://${authHost}${loginPath}`;
};
