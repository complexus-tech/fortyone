import ky, { Options } from "ky";
import { auth } from "@/auth";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

const createClient = async () => {
  // const session = await auth();
  // const prefixUrl = session
  //   ? `${apiURL}/workspaces/${session.activeWorkspace.id}/`
  //   : apiURL + "/";
  const prefixUrl = `${apiURL}/workspaces/3589aaa4-f1f4-40bb-ae1c-9104dd537d8c/`;

  return ky.create({
    prefixUrl,
    // hooks: {
    //   beforeRequest: [
    //     async (request) => {
    //       if (session) {
    //         // request.headers.set("Authorization", `Bearer ${session.token}`);
    //       }
    //     },
    //   ],
    // },
  });
};

export const get = async <T>(url: string, options?: Options) => {
  const client = await createClient();
  return client.get(url, options).json<T>();
};

export const post = async <T, U>(url: string, json: T, options?: Options) => {
  const client = await createClient();
  return client.post(url, { json, ...options }).json<U>();
};

export const put = async <T, U>(url: string, json: T, options?: Options) => {
  const client = await createClient();
  return client.put(url, { json, ...options }).json<U>();
};

export const patch = async <T, U>(url: string, json: T, options?: Options) => {
  const client = await createClient();
  return client.patch(url, { json, ...options }).json<U>();
};

export const remove = async <T>(url: string, options?: Options) => {
  const client = await createClient();
  return client.delete(url, options).json<T>();
};
