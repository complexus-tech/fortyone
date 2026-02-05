import { getApiUrl } from "@/lib/api-url";

const apiURL = getApiUrl();

export const exchangeSessionToken = async (token?: string) => {
  if (!token) return;

  try {
    await fetch(`${apiURL}/users/session`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
      credentials: "include",
    });
  } catch {
    // Exchange is best-effort; keep UI responsive if it fails.
  }
};
