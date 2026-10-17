import { useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";
import { get, post } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { chatKeys } from "./constants";
import type { AudioTranscript } from "./audio-attachment-player";

const TRANSCRIPTION_ERRORS: Record<string, string> = {
  interrupted: "Transcription was interrupted. You can try again.",
  access_unavailable: "Transcription is unavailable for this conversation.",
  invalid_audio: "This recording could not be read. Try recording it again.",
};

const transcriptionSchema = z.object({
  status: z.enum(["not_requested", "pending", "processing", "ready", "failed"]),
  text: z.string().optional(),
  error: z.string().optional(),
  updatedAt: z.string().optional(),
});
export type ChatTranscription = z.infer<typeof transcriptionSchema>;
const decode = (value: unknown) =>
  z.object({ data: transcriptionSchema }).parse(value).data;
const path = (conversationId: string, attachmentId: string) =>
  `chat/conversations/${conversationId}/attachments/${attachmentId}/transcription`;

const requestErrorMessage = (error: Error) => {
  const status = "status" in error ? Number(error.status) : 0;
  if (status === 403) return "Transcription is not available for your account.";
  if (status === 404) return "This audio is no longer available.";
  if (status === 429) return "Transcription is busy. Please try again shortly.";
  return "Could not load the transcript. Please try again.";
};

export const getAudioTranscription = (
  ctx: WorkspaceCtx,
  conversationId: string,
  attachmentId: string,
  signal?: AbortSignal,
): Promise<ChatTranscription> =>
  get(path(conversationId, attachmentId), ctx, { signal, retry: 0 }, decode);

export const requestAudioTranscription = (
  ctx: WorkspaceCtx,
  conversationId: string,
  attachmentId: string,
  retry: boolean,
): Promise<ChatTranscription> =>
  post(
    path(conversationId, attachmentId),
    { retry },
    ctx,
    { retry: 0 },
    decode,
  );

// Nothing is fetched for every audio row. The first explicit click checks for a
// saved result before requesting work; only an open pending request is polled.
export const useAudioTranscription = (
  conversationId: string,
  attachmentId: string,
) => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const client = useQueryClient();
  const [requestedScope, setRequestedScope] = useState<string | null>(null);
  const inFlight = useRef<string | null>(null);
  const queryKey = [
    ...chatKeys.all(workspaceSlug),
    "transcription",
    conversationId,
    attachmentId,
    session?.user.id,
  ];
  const cached = client.getQueryData<ChatTranscription>(queryKey);
  const scope = JSON.stringify(queryKey);
  const waiting =
    cached?.status === "pending" || cached?.status === "processing";
  const active = requestedScope === scope || waiting;
  const result = useQuery({
    queryKey,
    queryFn: ({ signal }) =>
      getAudioTranscription(
        { workspaceSlug },
        conversationId,
        attachmentId,
        signal,
      ),
    enabled: active && Boolean(session?.user.id),
    staleTime: 60_000,
    gcTime: 30 * 60_000,
    retry: false,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return active &&
        !query.state.error &&
        (status === "pending" || status === "processing")
        ? 2_500
        : false;
    },
    refetchIntervalInBackground: false,
  });
  const mutation = useMutation({
    mutationKey: queryKey,
    mutationFn: async (input: {
      workspaceSlug: string;
      conversationId: string;
      attachmentId: string;
      queryKey: typeof queryKey;
      scope: string;
    }) => {
      const existing = await client.fetchQuery({
        queryKey: input.queryKey,
        queryFn: ({ signal }) =>
          getAudioTranscription(
            { workspaceSlug: input.workspaceSlug },
            input.conversationId,
            input.attachmentId,
            signal,
          ),
        staleTime: 0,
        retry: false,
      });
      if (existing.status !== "not_requested" && existing.status !== "failed")
        return existing;
      return requestAudioTranscription(
        { workspaceSlug: input.workspaceSlug },
        input.conversationId,
        input.attachmentId,
        existing.status === "failed",
      );
    },
    onSuccess: (data, input) => {
      client.setQueryData(input.queryKey, data);
      setRequestedScope(input.scope);
    },
    onSettled: (_data, _error, input) => {
      if (inFlight.current === input.scope) inFlight.current = null;
    },
  });
  const data = result.data;
  const currentMutation = mutation.variables?.scope === scope;
  const error = (currentMutation ? mutation.error : null) ?? result.error;
  let transcript: AudioTranscript = { status: "idle" };
  if (currentMutation && mutation.isPending) transcript = { status: "pending" };
  else if (error)
    transcript = { status: "error", error: requestErrorMessage(error) };
  else if (data?.status === "pending" || data?.status === "processing")
    transcript = { status: "pending" };
  else if (data?.status === "ready")
    transcript = { status: "ready", text: data.text };
  else if (data?.status === "failed")
    transcript = {
      status: "error",
      error:
        TRANSCRIPTION_ERRORS[data.error ?? ""] ??
        "The transcript could not be created. You can try again.",
    };
  return {
    transcript,
    request: () => {
      if (
        !session?.user.id ||
        !conversationId ||
        !attachmentId ||
        inFlight.current === scope ||
        client.isMutating({ mutationKey: queryKey, exact: true }) > 0
      )
        return;
      inFlight.current = scope;
      mutation.mutate({
        workspaceSlug,
        conversationId,
        attachmentId,
        queryKey,
        scope,
      });
    },
  };
};
