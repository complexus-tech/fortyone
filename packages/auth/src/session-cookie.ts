export const SESSION_COOKIE_NAME = "fortyone_session";

export const getSessionCookieHeader = (
  cookieHeader: string | null,
): string | null => {
  if (!cookieHeader) return null;

  for (const rawCookie of cookieHeader.split(";")) {
    const separatorIndex = rawCookie.indexOf("=");
    if (separatorIndex < 0) continue;

    const name = rawCookie.slice(0, separatorIndex).trim();
    const value = rawCookie.slice(separatorIndex + 1).trim();

    if (name === SESSION_COOKIE_NAME && value.length > 0) {
      return `${SESSION_COOKIE_NAME}=${value}`;
    }
  }

  return null;
};
