import type { SignInTransaction } from "./auth-contract";
import * as SecureStore from "expo-secure-store";
import { createSessionStorage } from "./session-storage";

export { parseSessionCookie, SESSION_COOKIE_NAME } from "./auth-contract";
export type { StoredSession, SignInTransaction } from "./auth-contract";
const TRANSACTION_KEY = "fortyone.sign-in.v1";
const RANDOM_VALUE = /^[A-Za-z0-9_-]{43}$/;
const TRANSACTION_TTL = 20 * 60 * 1000;
const sessionStorage = createSessionStorage({
  getItem: SecureStore.getItemAsync,
  setItem: SecureStore.setItemAsync,
  removeItem: SecureStore.deleteItemAsync,
});
export const saveSession = sessionStorage.save;
export const getStoredSession = sessionStorage.get;
export const clearStoredSession = sessionStorage.clear;

let transactionQueue: Promise<unknown> = Promise.resolve();
const serializeTransaction = <T>(operation: () => Promise<T>): Promise<T> => {
  const result = transactionQueue.then(operation, operation);
  transactionQueue = result.catch(() => undefined);
  return result;
};
export const saveSignInTransaction = (transaction: SignInTransaction) =>
  serializeTransaction(() =>
    SecureStore.setItemAsync(TRANSACTION_KEY, JSON.stringify(transaction)),
  );
export const clearSignInTransaction = () =>
  serializeTransaction(() => SecureStore.deleteItemAsync(TRANSACTION_KEY));
export const consumeSignInTransaction = () =>
  serializeTransaction(async (): Promise<SignInTransaction> => {
    const serialized = await SecureStore.getItemAsync(TRANSACTION_KEY);
    await SecureStore.deleteItemAsync(TRANSACTION_KEY);
    let transaction: Partial<SignInTransaction> = {};
    try {
      transaction = serialized ? JSON.parse(serialized) : {};
    } catch {
      // A damaged transaction must start over, never authenticate.
    }
    if (
      !transaction ||
      typeof transaction.state !== "string" ||
      !RANDOM_VALUE.test(transaction.state) ||
      typeof transaction.verifier !== "string" ||
      !RANDOM_VALUE.test(transaction.verifier) ||
      typeof transaction.createdAt !== "number" ||
      transaction.createdAt > Date.now() ||
      Date.now() - transaction.createdAt > TRANSACTION_TTL
    ) {
      throw new Error("This sign-in attempt has expired. Please start again.");
    }
    return transaction as SignInTransaction;
  });
