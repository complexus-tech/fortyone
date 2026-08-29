export const isFigmaURL = (value: string) => {
  try {
    const hostname = new URL(value).hostname.toLowerCase();
    return hostname === "figma.com" || hostname === "www.figma.com";
  } catch {
    return false;
  }
};
