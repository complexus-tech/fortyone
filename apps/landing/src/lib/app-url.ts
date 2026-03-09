const DEFAULT_APP_URL = "https://cloud.fortyone.app";

export const APP_URL = (
  process.env.NEXT_PUBLIC_APP_URL ?? DEFAULT_APP_URL
).replace(/\/$/, "");

export const SIGNUP_URL = `${APP_URL}/signup`;
