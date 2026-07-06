import { getPublicEnv } from "@/public-env";

const stripTrailingSlash = (value: string) => value.replace(/\/+$/, "");

const isLocalHost = (host: string) =>
  host === "localhost" || host === "127.0.0.1" || host.endsWith(".localhost");

export const getApiUrl = () => {
  if (typeof window === "undefined") {
    return stripTrailingSlash(process.env.NEXT_PUBLIC_API_URL ?? "");
  }

  const configured = stripTrailingSlash(getPublicEnv().API_URL);

  try {
    const apiUrl = new URL(configured);
    const appHost = window.location.hostname;

    if (isLocalHost(apiUrl.hostname) && isLocalHost(appHost)) {
      apiUrl.hostname = appHost;
      return stripTrailingSlash(apiUrl.toString());
    }

    return configured;
  } catch {
    return configured;
  }
};

export const getProjectsUrl = () => {
  if (typeof window === "undefined") {
    return stripTrailingSlash(
      process.env.NEXT_PUBLIC_PROJECTS_URL ?? "http://localhost:3000",
    );
  }

  return stripTrailingSlash(getPublicEnv().PROJECTS_URL);
};
