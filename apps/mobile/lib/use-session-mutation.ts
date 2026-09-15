import type { DefaultError, QueryClient } from "@tanstack/react-query";
import { createContext, useContext } from "react";
import { useMutation } from "@tanstack/react-query";
import {
  withSessionMutationGuard,
  type SessionMutationAdmission,
  type SessionMutationOptions,
} from "./session-mutation";

export const SessionMutationContext = createContext<
  (SessionMutationAdmission & { client: QueryClient }) | null
>(null);

export function useSessionMutation<
  TData = unknown,
  TError = DefaultError,
  TVariables = void,
  TOnMutateResult = unknown,
>(options: SessionMutationOptions<TData, TError, TVariables, TOnMutateResult>) {
  const session = useContext(SessionMutationContext);
  if (!session)
    throw new Error("useSessionMutation requires SessionQueryProvider.");
  return useMutation(
    withSessionMutationGuard(session, options),
    session.client,
  );
}
