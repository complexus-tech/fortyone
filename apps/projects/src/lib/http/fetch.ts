import type { Options } from "ky";
import ky from "ky";
import { auth } from "@/auth";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

const createClient = async () => {
  const session = await auth();
  const prefixUrl = `${apiURL}/workspaces/${session?.activeWorkspace.id}/`;
  return ky.create({
    prefixUrl,
  });
};

const addAuth = async (options?: Options) => {
  const session = await auth();
  if (session) {
    return {
      headers: {
        Authorization: `Bearer ${session.token}`,
      },
      ...options,
    };
  }
  return options;
};

export const get = async <T>(url: string, options?: Options) => {
  const client = await createClient();
  const authenticatedOptions = await addAuth(options);
  return client.get(url, authenticatedOptions).json<T>();
};

export const post = async <T, U>(url: string, json: T, options?: Options) => {
  const client = await createClient();
  const authenticatedOptions = await addAuth(options);
  return client.post(url, { json, ...authenticatedOptions }).json<U>();
};

export const put = async <T, U>(url: string, json: T, options?: Options) => {
  const client = await createClient();
  const authenticatedOptions = await addAuth(options);
  return client.put(url, { json, ...authenticatedOptions }).json<U>();
};

export const patch = async <T, U>(url: string, json: T, options?: Options) => {
  const client = await createClient();
  const authenticatedOptions = await addAuth(options);
  return client.patch(url, { json, ...authenticatedOptions }).json<U>();
};

export const remove = async <T>(url: string, options?: Options) => {
  const client = await createClient();
  const authenticatedOptions = await addAuth(options);
  return client.delete(url, authenticatedOptions).json<T>();
};
