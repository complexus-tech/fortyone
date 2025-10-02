import ky from "ky";
import { ApiError } from "./error";
import { useAuthStore } from "@/store/auth";

const apiURL = process.env.EXPO_PUBLIC_API_URL;

// Create HTTP client with workspace context
const createClient = (useWorkspace = true) => {
  const { token, workspace } = useAuthStore.getState();

  if (!token) throw new Error("No authentication token found");

  const prefixUrl =
    useWorkspace && workspace
      ? `${apiURL}/workspaces/${workspace}/`
      : `${apiURL}/`;

  return ky.create({
    prefixUrl,
    headers: { Authorization: `Bearer ${token}` },
    hooks: {
      beforeError: [
        async (error) => {
          const { response } = error;
          if (response.body) {
            const data = await response.json();
            throw new ApiError(error.message, response.status, data);
          }
          throw error;
        },
      ],
    },
  });
};

// HTTP methods
export const get = async <T>(
  url: string,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const client = createClient(options?.useWorkspace);
  return client.get(url, { headers: options?.headers }).json<T>();
};

export const post = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const client = createClient(options?.useWorkspace);

  if (json instanceof FormData) {
    return client
      .post(url, { body: json, headers: options?.headers })
      .json<U>();
  }

  return client.post(url, { json, headers: options?.headers }).json<U>();
};

export const put = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const client = createClient(options?.useWorkspace);

  if (json instanceof FormData) {
    return client.put(url, { body: json, headers: options?.headers }).json<U>();
  }

  return client.put(url, { json, headers: options?.headers }).json<U>();
};

export const patch = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const client = createClient(options?.useWorkspace);

  if (json instanceof FormData) {
    return client
      .patch(url, { body: json, headers: options?.headers })
      .json<U>();
  }

  return client.patch(url, { json, headers: options?.headers }).json<U>();
};

export const remove = async <T>(
  url: string,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const client = createClient(options?.useWorkspace);
  return client.delete(url, { headers: options?.headers }).json<T>();
};
