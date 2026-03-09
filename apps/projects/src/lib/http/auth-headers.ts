export type AuthHeaderOptions = {
  token?: string;
  cookieHeader?: string;
};

export const buildAuthHeaders = ({ cookieHeader }: AuthHeaderOptions = {}) => {
  const headers: Record<string, string> = {};

  if (cookieHeader) {
    headers.Cookie = cookieHeader;
  }

  return headers;
};
