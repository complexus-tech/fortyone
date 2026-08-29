const LOOPBACK_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);

export const normalizeFeedbackWidgetOrigin = (value: string) => {
  const input = value.trim();
  if (!input || input.includes("*")) {
    throw new Error("Enter one exact origin without wildcards");
  }

  let url: URL;
  try {
    url = new URL(input);
  } catch {
    throw new Error("Enter a complete origin such as https://app.example.com");
  }

  const isLocal = LOOPBACK_HOSTS.has(url.hostname);
  if (url.protocol !== "https:" && !(url.protocol === "http:" && isLocal)) {
    throw new Error("Origins must use HTTPS (HTTP is allowed for localhost)");
  }
  if (
    url.username ||
    url.password ||
    (url.pathname !== "/" && url.pathname !== "") ||
    url.search ||
    url.hash
  ) {
    throw new Error("Use only the scheme, host, and optional port");
  }

  return url.origin;
};

export const normalizeFeedbackWidgetOrigins = (values: string[]) => {
  const normalized = values.map(normalizeFeedbackWidgetOrigin);
  if (new Set(normalized).size !== normalized.length) {
    throw new Error("Each allowed origin can only be added once");
  }
  return normalized;
};
