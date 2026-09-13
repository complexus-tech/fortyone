const localHostname = (hostname: string) =>
  hostname === "localhost" ||
  hostname === "127.0.0.1" ||
  hostname.endsWith(".localhost");

const configuredURL = (value: string | undefined, name: string) => {
  if (!value) throw new Error(`${name} is not configured.`);
  const url = new URL(value);
  if (
    url.username ||
    url.password ||
    url.search ||
    url.hash ||
    (url.protocol !== "https:" &&
      !(
        typeof __DEV__ !== "undefined" &&
        __DEV__ &&
        url.protocol === "http:" &&
        localHostname(url.hostname)
      ))
  ) {
    throw new Error(`${name} must be a trusted HTTPS URL.`);
  }
  return url;
};

export const getApiURL = () =>
  configuredURL(process.env.EXPO_PUBLIC_API_URL, "EXPO_PUBLIC_API_URL");
export const getApplicationURL = () =>
  configuredURL(
    process.env.EXPO_PUBLIC_APP_URL ?? "https://cloud.fortyone.app",
    "EXPO_PUBLIC_APP_URL",
  );
export const assertApiRequestURL = (requestURL: string) => {
  const base = getApiURL();
  const target = new URL(requestURL);
  const basePath = `${base.pathname.replace(/\/$/, "")}/`;
  if (
    target.origin !== base.origin ||
    target.username ||
    target.password ||
    !target.pathname.startsWith(basePath)
  ) {
    throw new Error("The API client cannot send credentials to this URL.");
  }
};
