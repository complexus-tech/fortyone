import type {
  DefaultError,
  MutationFunction,
  UseMutationOptions,
} from "@tanstack/react-query";

export type SessionMutationAdmission = {
  assertCanMutate: () => void;
  isActive: () => boolean;
};

export type SessionMutationOptions<TData, TError, TVariables, TOnMutateResult> =
  UseMutationOptions<TData, TError, TVariables, TOnMutateResult> & {
    mutationFn: MutationFunction<TData, TVariables>;
  };

/** Capture the originating provider; never admit a write using the next session. */
export function withSessionMutationGuard<
  TData = unknown,
  TError = DefaultError,
  TVariables = void,
  TOnMutateResult = unknown,
>(
  session: SessionMutationAdmission,
  options: SessionMutationOptions<TData, TError, TVariables, TOnMutateResult>,
): UseMutationOptions<TData, TError, TVariables, TOnMutateResult> {
  return {
    ...options,
    mutationFn: (variables, context) => {
      // TanStack awaits onMutate before invoking this function. Admission here
      // covers cancellation/optimistic work, queued mutations, and retries.
      session.assertCanMutate();
      return options.mutationFn(variables, context);
    },
    onSuccess: (...args) =>
      session.isActive() ? options.onSuccess?.(...args) : undefined,
    onError: (...args) =>
      session.isActive() ? options.onError?.(...args) : undefined,
    onSettled: (...args) =>
      session.isActive() ? options.onSettled?.(...args) : undefined,
  };
}
