import type { QueryKey } from "@tanstack/react-query";
import { useQueryClient } from "@tanstack/react-query";
import { useRef } from "react";
import { useSessionMutation } from "@/lib/use-session-mutation";

export function useDetailMutation<Input, Output>(
  mutationFn: (input: Input) => Promise<Output>,
  queryKeys: QueryKey[],
) {
  const client = useQueryClient();
  const pending = useRef(false);
  const mutation = useSessionMutation({
    mutationFn,
    retry: false,
    onSuccess: () => {
      // Cache refresh errors must never turn a completed write into a retryable send.
      for (const queryKey of queryKeys)
        void client.invalidateQueries({ queryKey });
    },
  });
  return {
    ...mutation,
    execute: async (input: Input) => {
      if (pending.current)
        throw new Error("Please wait for the current change to finish.");
      pending.current = true;
      return mutation.mutateAsync(input).finally(() => {
        pending.current = false;
      });
    },
  };
}
