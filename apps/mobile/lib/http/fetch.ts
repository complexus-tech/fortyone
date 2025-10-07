import ky from "ky";
import { useAuthStore } from "@/store/auth";
import { ApiResponse } from "@/types";

const apiURL = process.env.EXPO_PUBLIC_API_URL;

// Create HTTP client with workspace context
const createClient = (useWorkspace = true) => {
  const [token, workspace] = [
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI4YTc5ODExMi05MGZlLTQ5NWUtOWYxYy1mMzY2NTVlM2Q4YWIiLCJleHAiOjE3NjI4ODMyMDMsIm5iZiI6MTc1OTQyNzIwMywiaWF0IjoxNzU5NDI3MjAzfQ.K05W85tEEWQ5dFqu7bgXjjowkk_zYowwKSJ_VMXR7_o",
    "complexus",
  ];

  const prefixUrl =
    useWorkspace && workspace
      ? `${apiURL}/workspaces/${workspace}/`
      : `${apiURL}/`;

  return ky.create({
    prefixUrl,
    headers: { Authorization: `Bearer ${token}` },
    retry: 0,
    hooks: {
      beforeError: [
        async (error) => {
          const { response } = error;
          const data = await response.json<ApiResponse<null>>();
          const errorMessage = data?.error?.message || "An error occurred";
          throw new Error(errorMessage);
        },
      ],
    },
  });
};

// HTTP methods
export const get = async <T>(
  url: string,
  options?: { useWorkspace?: boolean }
) => {
  const client = createClient(options?.useWorkspace);
  return client.get(url).json<T>();
};

export const post = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean }
) => {
  const client = createClient(options?.useWorkspace);

  if (json instanceof FormData) {
    return client.post(url, { body: json }).json<U>();
  }

  return client.post(url, { json }).json<U>();
};

export const put = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean }
) => {
  const client = createClient(options?.useWorkspace);

  if (json instanceof FormData) {
    return client.put(url, { body: json }).json<U>();
  }

  return client.put(url, { json }).json<U>();
};

export const patch = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean }
) => {
  const client = createClient(options?.useWorkspace);

  if (json instanceof FormData) {
    return client.patch(url, { body: json }).json<U>();
  }

  return client.patch(url, { json }).json<U>();
};

export const remove = async <T>(
  url: string,
  options?: { useWorkspace?: boolean }
) => {
  const client = createClient(options?.useWorkspace);
  return client.delete(url).json<T>();
};
