import type { ApiResponse } from "@/types";
import type { Options } from "ky";
import ky from "ky";
import { fetch as expoFetch } from "expo/fetch";
import { useAuthStore } from "@/store/auth";
import { getStoredSession } from "@/lib/auth";
import { assertApiRequestURL, getApiURL, getApplicationURL } from "./config";

export class HttpError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "HttpError";
  }
}
export type RequestOptions = {
  useWorkspace?: boolean;
  signal?: AbortSignal;
  // null makes an unauthenticated exchange; a captured cookie permits remote
  // revocation after the local credential has already been deleted.
  sessionCookie?: string | null;
  handleUnauthorized?: boolean;
};

export const request = async (
  method: "get" | "post" | "put" | "patch" | "delete",
  path: string,
  payload?: unknown,
  options: RequestOptions = {},
) => {
  const apiURL = getApiURL().toString().replace(/\/$/, "");
  const { workspace, sessionEpoch } = useAuthStore.getState();
  const session =
    options.sessionCookie === undefined ? await getStoredSession() : null;
  if (session && session.apiOrigin !== getApiURL().origin) {
    await useAuthStore.getState().expireSession(session.cookie);
    throw new HttpError("Sign in again for this API environment.", 401);
  }
  const cookie =
    options.sessionCookie === undefined
      ? session?.cookie
      : options.sessionCookie;
  if (
    options.sessionCookie === undefined &&
    !cookie &&
    useAuthStore.getState().isAuthenticated &&
    useAuthStore.getState().sessionEpoch === sessionEpoch
  ) {
    await useAuthStore.getState().expireSession("");
    throw new HttpError("Your session expired. Please sign in again.", 401);
  }
  const prefixUrl =
    options.useWorkspace !== false && workspace
      ? `${apiURL}/workspaces/${encodeURIComponent(workspace)}/`
      : `${apiURL}/`;
  const client = ky.create({
    prefixUrl,
    fetch: expoFetch as typeof globalThis.fetch,
    // SDK 57 honors omit on iOS and Android. Do not let native cookie jars undo
    // local logout or mix an older account into a new session.
    credentials: "omit",
    redirect: "error",
    retry: 0,
    timeout: 20_000,
    hooks: {
      beforeRequest: [
        (outgoing) => {
          assertApiRequestURL(outgoing.url);
          outgoing.headers.set("Origin", getApplicationURL().origin);
          if (cookie) outgoing.headers.set("Cookie", cookie);
          if (
            options.sessionCookie === undefined &&
            useAuthStore.getState().sessionEpoch !== sessionEpoch
          ) {
            throw new Error(
              "The session changed before this request could be sent.",
            );
          }
        },
      ],
      beforeError: [
        async (error) => {
          const response = error.response;
          const data = (await response
            .clone()
            .json()
            .catch(() => null)) as ApiResponse<null> | null;
          if (
            response.status === 401 &&
            cookie &&
            options.handleUnauthorized !== false
          ) {
            await useAuthStore.getState().expireSession(cookie);
          }
          throw new HttpError(
            data?.error?.message ?? "The request could not be completed.",
            response.status,
          );
        },
      ],
    },
  });
  const body: Options =
    payload instanceof FormData
      ? { body: payload }
      : payload === undefined
        ? {}
        : { json: payload };
  return client[method](path, { ...body, signal: options.signal });
};
const readJSON = async <T>(response: Response): Promise<T> =>
  response.status === 204 ? (undefined as T) : (response.json() as Promise<T>);
export const get = async <T>(url: string, options?: RequestOptions) =>
  readJSON<T>(await request("get", url, undefined, options));
export const post = async <T, U>(
  url: string,
  json: T,
  options?: RequestOptions,
) => readJSON<U>(await request("post", url, json, options));
export const put = async <T, U>(
  url: string,
  json: T,
  options?: RequestOptions,
) => readJSON<U>(await request("put", url, json, options));
export const patch = async <T, U>(
  url: string,
  json: T,
  options?: RequestOptions,
) => readJSON<U>(await request("patch", url, json, options));
export const remove = async <T>(url: string, options?: RequestOptions) =>
  readJSON<T>(await request("delete", url, undefined, options));
