export type AuthHeaderOptions = {
  token?: string;
  cookieHeader?: string;
};

const SESSION_COOKIE_NAME = "fortyone_session";

const hasSessionCookie = (cookieHeader?: string) =>
  Boolean(cookieHeader?.includes(`${SESSION_COOKIE_NAME}=`));

export const buildAuthHeaders = ({
  token,
  cookieHeader,
}: AuthHeaderOptions = {}) => {
  const headers: Record<string, string> = {};

  if (cookieHeader) {
    headers.Cookie = cookieHeader;
  }

  const shouldUseToken =
    typeof window === "undefined" && token && !hasSessionCookie(cookieHeader);

  if (shouldUseToken) {
    headers.Authorization = `Bearer ${token}`;
  }

  return headers;
};
