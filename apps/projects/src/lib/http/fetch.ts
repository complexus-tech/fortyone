"use server";
import type { Options } from "ky";
import ky from "ky";
import { headers } from "next/headers";
import { auth } from "@/auth";
// import { getWorkspaces } from "../queries/workspaces/get-workspaces";
import { ApiError } from "./error";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

const createClient = async () => {
  const headersList = await headers();
  const subdomain = headersList.get("host")!.split(".")[0];
  const session = await auth();

  const workspaces = session?.workspaces || [];
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
  );

  const prefixUrl = `${apiURL}/workspaces/${workspace?.id}/`;
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
