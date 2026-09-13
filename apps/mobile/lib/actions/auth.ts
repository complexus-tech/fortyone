import type { ApiResponse, User } from "@/types";
import * as Crypto from "expo-crypto";
import { post, request } from "@/lib/http";
import { getApiURL, getApplicationURL } from "@/lib/http/config";
import {
  consumeSignInTransaction,
  parseSessionCookie,
  saveSession,
  saveSignInTransaction,
} from "@/lib/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import {
  MOBILE_REDIRECT_URI,
  readMobileCallbackCode,
} from "@/lib/auth-contract";
import { useAuthStore } from "@/store/auth";

export { MOBILE_REDIRECT_URI } from "@/lib/auth-contract";
let pendingExchange: {
  url: string;
  promise: Promise<{ userId: string; workspace: string }>;
} | null = null;
const toBase64URL = (base64: string) =>
  base64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
const randomValue = async () =>
  toBase64URL(
    btoa(String.fromCharCode(...(await Crypto.getRandomBytesAsync(32)))),
  );

export const beginMobileSignIn = async () => {
  pendingExchange = null;
  const [state, verifier] = await Promise.all([randomValue(), randomValue()]);
  const challenge = toBase64URL(
    await Crypto.digestStringAsync(
      Crypto.CryptoDigestAlgorithm.SHA256,
      verifier,
      {
        encoding: Crypto.CryptoEncoding.BASE64,
      },
    ),
  );
  await saveSignInTransaction({ state, verifier, createdAt: Date.now() });
  const url = new URL("/auth/mobile", getApplicationURL());
  url.searchParams.set("state", state);
  url.searchParams.set("code_challenge", challenge);
  return url.toString();
};

const exchangeCode = async (callbackURL: string) => {
  const sessionEpoch = useAuthStore.getState().sessionEpoch;
  const isCurrent = () => {
    const state = useAuthStore.getState();
    return !state.isLoading && state.sessionEpoch === sessionEpoch;
  };
  const transaction = await consumeSignInTransaction();
  const code = readMobileCallbackCode(callbackURL, transaction.state);
  const response = await request(
    "post",
    "auth/mobile/exchange",
    {
      code,
      state: transaction.state,
      codeVerifier: transaction.verifier,
      redirectUri: MOBILE_REDIRECT_URI,
    },
    { useWorkspace: false, sessionCookie: null, handleUnauthorized: false },
  );
  const result = (await response.json()) as ApiResponse<User>;
  if (!result.data?.id) throw new Error("Sign-in did not return an account.");
  const session = {
    ...parseSessionCookie(response.headers.get("set-cookie"), getApiURL()),
    apiOrigin: getApiURL().origin,
    userId: result.data.id,
    workspace: null as string | null,
  };
  if (!isCurrent()) throw new Error("Sign-in was cancelled.");
  await saveSession(session, isCurrent);
  const workspaces = await getWorkspaces();
  const activeWorkspace =
    workspaces.find(
      (workspace) => workspace.id === result.data?.lastUsedWorkspaceId,
    ) ?? workspaces[0];
  if (!activeWorkspace)
    throw new Error(
      "This account has no workspace. Complete workspace setup and try again.",
    );
  session.workspace = activeWorkspace.slug;
  if (!isCurrent()) throw new Error("Sign-in was cancelled.");
  await saveSession(session, isCurrent);
  return { userId: session.userId, workspace: session.workspace };
};

// Browser completion and the Router deep link can arrive together. Redeem a
// callback once, including when the app resumes after being terminated.
export const authenticateWithCode = (callbackURL: string) => {
  if (pendingExchange?.url === callbackURL) return pendingExchange.promise;
  const promise = exchangeCode(callbackURL);
  pendingExchange = { url: callbackURL, promise };
  return promise;
};

export const switchWorkspace = async (workspaceId: string) => {
  const response = await post<{ workspaceId: string }, ApiResponse<User>>(
    "workspaces/switch",
    { workspaceId },
    { useWorkspace: false },
  );
  return response.data?.lastUsedWorkspaceId;
};
