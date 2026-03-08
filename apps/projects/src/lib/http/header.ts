export const getHeaders = async () => {
  if (typeof window !== "undefined") {
    return new Headers();
  }

  const { headers } = await import("next/headers");
  return headers();
};

export const getCookieHeader = async () => {
  const headersList = await getHeaders();
  return headersList.get("cookie") ?? "";
};
