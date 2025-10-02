import ky from "ky";
import { ApiError } from "./error";

const apiURL = process.env.EXPO_PUBLIC_API_URL;

// Get stored auth data
const getAuthData = async () => {
  const [token, workspaceId] = ["test", "test"];
  return { token, workspaceId };
};

// Create HTTP client with workspace context
const createClient = (token: string, workspaceId?: string) => {
  const prefixUrl = workspaceId
    ? `${apiURL}/workspaces/${workspaceId}/`
    : `${apiURL}/`;

  return ky.create({
    prefixUrl,
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
  const { token, workspaceId } = await getAuthData();
  if (!token) throw new Error("No authentication token found");

  const client = createClient(
    token,
    options?.useWorkspace !== false ? workspaceId : undefined
  );
  const mergedOptions = {
    headers: {
      Authorization: `Bearer ${token}`,
      ...(options?.headers || {}),
    },
  };

  return client.get(url, mergedOptions).json<T>();
};

export const post = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const { token, workspaceId } = await getAuthData();
  if (!token) throw new Error("No authentication token found");

  const client = createClient(
    token,
    options?.useWorkspace !== false ? workspaceId : undefined
  );
  const mergedOptions = {
    headers: {
      Authorization: `Bearer ${token}`,
      ...(options?.headers || {}),
    },
  };

  if (json instanceof FormData) {
    return client.post(url, { body: json, ...mergedOptions }).json<U>();
  }

  return client.post(url, { json, ...mergedOptions }).json<U>();
};

export const put = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const { token, workspaceId } = await getAuthData();
  if (!token) throw new Error("No authentication token found");

  const client = createClient(
    token,
    options?.useWorkspace !== false ? workspaceId : undefined
  );
  const mergedOptions = {
    headers: {
      Authorization: `Bearer ${token}`,
      ...(options?.headers || {}),
    },
  };

  if (json instanceof FormData) {
    return client.put(url, { body: json, ...mergedOptions }).json<U>();
  }
  return client.put(url, { json, ...mergedOptions }).json<U>();
};

export const patch = async <T, U>(
  url: string,
  json: T,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const { token, workspaceId } = await getAuthData();
  if (!token) throw new Error("No authentication token found");

  const client = createClient(
    token,
    options?.useWorkspace !== false ? workspaceId : undefined
  );
  const mergedOptions = {
    headers: {
      Authorization: `Bearer ${token}`,
      ...(options?.headers || {}),
    },
  };

  if (json instanceof FormData) {
    return client.patch(url, { body: json, ...mergedOptions }).json<U>();
  }
  return client.patch(url, { json, ...mergedOptions }).json<U>();
};

export const remove = async <T>(
  url: string,
  options?: { useWorkspace?: boolean; headers?: Record<string, string> }
) => {
  const { token, workspaceId } = await getAuthData();
  if (!token) throw new Error("No authentication token found");

  const client = createClient(
    token,
    options?.useWorkspace !== false ? workspaceId : undefined
  );
  const mergedOptions = {
    headers: {
      Authorization: `Bearer ${token}`,
      ...(options?.headers || {}),
    },
  };

  return client.delete(url, mergedOptions).json<T>();
};
