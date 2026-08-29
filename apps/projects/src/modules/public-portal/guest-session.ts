import "server-only";

import { createHash } from "node:crypto";
import { cookies } from "next/headers";

const FEEDBACK_SESSION_COOKIE_PREFIX = "fortyone_feedback_session";
const FEEDBACK_PREFERENCE_COOKIE_PREFIX = "fortyone_feedback_preferences";

const getPortalCookieSuffix = (portalSlug: string) =>
  createHash("sha256")
    .update(portalSlug.trim().toLowerCase())
    .digest("hex")
    .slice(0, 20);

const getFeedbackSessionCookieName = (portalSlug: string) =>
  `${FEEDBACK_SESSION_COOKIE_PREFIX}_${getPortalCookieSuffix(portalSlug)}`;

const getFeedbackPreferenceCookieName = (portalSlug: string) =>
  `${FEEDBACK_PREFERENCE_COOKIE_PREFIX}_${getPortalCookieSuffix(portalSlug)}`;

const getCookieOptions = (expiresAt: string) => {
  const expires = new Date(expiresAt);
  if (Number.isNaN(expires.getTime())) {
    throw new Error("Feedback session expiry is invalid");
  }

  return {
    expires,
    httpOnly: true,
    path: "/",
    sameSite: "lax" as const,
    secure: process.env.NODE_ENV === "production",
  };
};

export const getFeedbackSessionToken = async (portalSlug: string) =>
  (await cookies()).get(getFeedbackSessionCookieName(portalSlug))?.value ??
  null;

export const setFeedbackSessionToken = async ({
  expiresAt,
  portalSlug,
  token,
}: {
  expiresAt: string;
  portalSlug: string;
  token: string;
}) => {
  const cookieStore = await cookies();
  cookieStore.set(
    getFeedbackSessionCookieName(portalSlug),
    token,
    getCookieOptions(expiresAt),
  );
};

export const clearFeedbackSessionToken = async (portalSlug: string) => {
  const cookieStore = await cookies();
  cookieStore.delete(getFeedbackSessionCookieName(portalSlug));
};

export const getFeedbackPreferenceSessionToken = async (portalSlug: string) =>
  (await cookies()).get(getFeedbackPreferenceCookieName(portalSlug))?.value ??
  null;

export const setFeedbackPreferenceSessionToken = async ({
  expiresAt,
  portalSlug,
  token,
}: {
  expiresAt: string;
  portalSlug: string;
  token: string;
}) => {
  const cookieStore = await cookies();
  cookieStore.set(
    getFeedbackPreferenceCookieName(portalSlug),
    token,
    getCookieOptions(expiresAt),
  );
};

export const getFeedbackSessionAuthorization = async (portalSlug: string) => {
  const token = await getFeedbackSessionToken(portalSlug);
  return token ? `FeedbackSession ${token}` : null;
};

export const getFeedbackPreferenceAuthorization = async (
  portalSlug: string,
) => {
  const preferenceToken = await getFeedbackPreferenceSessionToken(portalSlug);
  if (preferenceToken) return `PreferenceSession ${preferenceToken}`;

  return getFeedbackSessionAuthorization(portalSlug);
};
