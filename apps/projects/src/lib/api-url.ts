import { getPublicEnv } from "@/public-env";

const isLocalHost = (host: string) =>
  host === "localhost" || host === "127.0.0.1" || host.endsWith(".localhost");

const stripTrailingSlash = (value: string) => value.replace(/\/+$/, "");

export const getApiUrl = () => {
  if (typeof window === "undefined") {
    return stripTrailingSlash(process.env.NEXT_PUBLIC_API_URL ?? "");
  }

  const configured = getPublicEnv().API_URL;
  const normalizedConfigured = stripTrailingSlash(configured);

  try {
    const apiUrl = new URL(normalizedConfigured);
    const appHost = window.location.hostname;

    // Keep local development hosts aligned so auth cookies are scoped to the same host.
    if (isLocalHost(apiUrl.hostname) && isLocalHost(appHost)) {
      apiUrl.hostname = appHost;
      return stripTrailingSlash(apiUrl.toString());
    }

    return normalizedConfigured;
  } catch {
    return normalizedConfigured;
  }
};
