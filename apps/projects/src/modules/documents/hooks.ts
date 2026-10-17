"use client";

import type { QueryClient } from "@tanstack/react-query";
import {
  useIsMutating,
  useMutation,
  useMutationState,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { documentKeys } from "@/shared/documents/keys";
import { documentMayaKey } from "./document-comment-maya";
import {
  addDocumentRelationshipAction,
  archiveDocumentAction,
  createDocumentAction,
  createDocumentCommentAction,
  deleteDocumentAction,
  duplicateDocumentAction,
  removeDocumentRelationshipAction,
  replyToDocumentCommentAction,
  resolveDocumentCommentAction,
  restoreArchivedDocumentAction,
  updateDocumentAccessAction,
  updateDocumentAction,
} from "./actions";
import {
  getDocument,
  getDocumentComments,
  getDocuments,
  getRelatedDocuments,
} from "./queries";
import type {
  DocumentAccessUpdate,
  DocumentComment,
  DocumentCommentThread,
  DocumentCreate,
  DocumentRelationType,
  DocumentUpdate,
  WorkspaceDocument,
} from "./types";

export const useDocuments = (search = "", scope = "all", limit?: number) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: documentKeys.list(workspaceSlug, search, scope, limit),
    queryFn: () =>
      getDocuments({ session: session!, workspaceSlug }, search, scope, limit),
    enabled: Boolean(session),
  });
};

export const useDocument = (documentId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: documentKeys.detail(workspaceSlug, documentId),
    queryFn: () =>
      getDocument(documentId, { session: session!, workspaceSlug }),
    enabled: Boolean(session && documentId),
  });
};

export const useDocumentComments = (documentId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const queryKey = documentKeys.comments(workspaceSlug, documentId);
  const mutationKey = documentCommentMutationKey(workspaceSlug, documentId);
  const pendingChanges = useIsMutating({ mutationKey, exact: true });
  return useQuery({
    queryKey,
    queryFn: () => {
      // Explicit refreshes bypass enabled, so protect optimistic comments here too.
      if (queryClient.isMutating({ mutationKey, exact: true })) {
        return (
          queryClient.getQueryData<DocumentCommentThread[]>(queryKey) ?? []
        );
      }
      return getDocumentComments(documentId, {
        session: session!,
        workspaceSlug,
      });
    },
    enabled: Boolean(session && documentId) && !pendingChanges,
  });
};

const documentCommentMutationKey = (
  workspaceSlug: string,
  documentId: string,
) => [...documentKeys.comments(workspaceSlug, documentId), "mutation"] as const;

const isOptimisticThread = (threadId: string) =>
  threadId.startsWith("optimistic-thread-");

const pendingCommentTargets = new WeakMap<QueryClient, Set<string>>();

const prepareDocumentCommentMutation = async (
  client: QueryClient,
  queryKey: ReturnType<typeof documentKeys.comments>,
  target: string,
  conflictMessage: string,
) => {
  const locks = pendingCommentTargets.get(client) ?? new Set<string>();
  const key = JSON.stringify([...queryKey, target]);
  if (locks.has(key)) throw new Error(conflictMessage);
  locks.add(key);
  pendingCommentTargets.set(client, locks);
  const release = () => {
    locks.delete(key);
  };
  try {
    await client.cancelQueries({ queryKey, exact: true });
    return release;
  } catch (error) {
    release();
    throw error;
  }
};

const prepareDocumentThreadMutation = (
  client: QueryClient,
  queryKey: ReturnType<typeof documentKeys.comments>,
  threadId: string,
) => {
  if (isOptimisticThread(threadId))
    throw new Error("Wait for this comment to finish saving");
  return prepareDocumentCommentMutation(
    client,
    queryKey,
    `thread:${threadId}`,
    "This thread is already being updated",
  );
};

const reconcileDocumentReply = (
  comments: DocumentComment[],
  saved: DocumentComment,
  optimisticId?: string,
) => {
  const remaining = comments.filter((comment) => comment.id !== optimisticId);
  const root = remaining.at(0);
  if (!root) return [saved];
  const replies = remaining.slice(1);
  const reconciled = remaining.some((comment) => comment.id === saved.id)
    ? replies
    : [...replies, saved];
  reconciled.sort(
    (left, right) =>
      Date.parse(left.createdAt) - Date.parse(right.createdAt) ||
      left.id.localeCompare(right.id),
  );
  return [root, ...reconciled];
};

export const useDocumentCommentMutations = (documentId: string) => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = documentKeys.comments(workspaceSlug, documentId);
  const mutationKey = documentCommentMutationKey(workspaceSlug, documentId);
  const pendingMutations = useMutationState({
    filters: { mutationKey, exact: true, status: "pending" },
    select: (mutation) =>
      mutation.state.variables as
        | { threadId?: string; draftTarget?: string; quote?: string }
        | undefined,
  });
  const invalidate = () => {
    if (queryClient.isMutating({ mutationKey, exact: true }) > 1) return;
    void queryClient.invalidateQueries({
      queryKey: documentMayaKey(workspaceSlug, documentId),
      exact: true,
    });
    return queryClient.invalidateQueries({ queryKey, exact: true });
  };
  const optimisticId = (kind: "comment" | "thread") =>
    `optimistic-${kind}-${crypto.randomUUID()}`;
  const optimisticComment = (body: string): DocumentComment => {
    const timestamp = new Date().toISOString();
    return {
      id: optimisticId("comment"),
      body,
      createdBy: session?.user.id ?? "",
      authorName: session?.user.name ?? "You",
      authorAvatar: session?.user.image ?? null,
      createdAt: timestamp,
      updatedAt: timestamp,
    };
  };
  const create = useMutation({
    mutationKey,
    mutationFn: async ({
      body,
      quote,
      anchorStart,
      anchorEnd,
      creationId,
    }: {
      body: string;
      creationId?: string;
      quote: string;
      anchorStart: number;
      anchorEnd: number;
      draftTarget?: string;
    }) => {
      const response = await createDocumentCommentAction(
        documentId,
        { body, quote, anchorStart, anchorEnd, creationId },
        workspaceSlug,
      );
      if (response.error || !response.data)
        throw new Error(
          response.error?.message ?? "The comment was not returned",
        );
      return response.data;
    },
    onMutate: async (payload) => {
      let release: (() => void) | undefined;
      if (payload.draftTarget) {
        release = await prepareDocumentCommentMutation(
          queryClient,
          queryKey,
          `draft:${payload.draftTarget}`,
          "This comment is already being sent",
        );
      } else {
        await queryClient.cancelQueries({ queryKey, exact: true });
      }
      const comment = optimisticComment(payload.body);
      const thread: DocumentCommentThread = {
        id: optimisticId("thread"),
        documentId,
        quote: payload.quote,
        anchorStart: payload.anchorStart,
        anchorEnd: payload.anchorEnd,
        createdBy: session?.user.id ?? "",
        resolvedAt: null,
        resolvedBy: null,
        createdAt: comment.createdAt,
        comments: [comment],
      };
      queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) => [
        thread,
        ...(threads ?? []),
      ]);
      return { optimisticThreadId: thread.id, release };
    },
    onError: (error, _payload, context) => {
      if (context) {
        queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
          threads?.filter((thread) => thread.id !== context.optimisticThreadId),
        );
      }
      toast.error("Could not add the comment", {
        description: error.message,
      });
    },
    onSuccess: (thread, _payload, context) => {
      queryClient.setQueryData<DocumentCommentThread[]>(
        queryKey,
        (threads = []) => {
          if (threads.some((cached) => cached.id === thread.id)) {
            return threads.filter(
              (cached) => cached.id !== context.optimisticThreadId,
            );
          }
          if (
            !threads.some((cached) => cached.id === context.optimisticThreadId)
          ) {
            return [thread, ...threads];
          }
          return threads.map((cached) =>
            cached.id === context.optimisticThreadId ? thread : cached,
          );
        },
      );
    },
    onSettled: (_data, _error, _payload, context) => {
      context?.release?.();
      return invalidate();
    },
  });
  const reply = useMutation({
    mutationKey,
    mutationFn: async ({
      threadId,
      body,
      creationId,
      replyToCommentId,
    }: {
      threadId: string;
      body: string;
      creationId?: string;
      replyToCommentId?: string;
    }) => {
      const response = await replyToDocumentCommentAction(
        documentId,
        threadId,
        body,
        workspaceSlug,
        { creationId, replyToCommentId },
      );
      if (response.error || !response.data)
        throw new Error(
          response.error?.message ?? "The reply was not returned",
        );
      return response.data;
    },
    onMutate: async ({ body, threadId }) => {
      const release = await prepareDocumentThreadMutation(
        queryClient,
        queryKey,
        threadId,
      );
      const comment = optimisticComment(body);
      queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
        threads?.map((thread) =>
          thread.id === threadId
            ? { ...thread, comments: [...thread.comments, comment] }
            : thread,
        ),
      );
      return { optimisticCommentId: comment.id, threadId, release };
    },
    onError: (error, _payload, context) => {
      if (context) {
        queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
          threads?.map((thread) =>
            thread.id === context.threadId
              ? {
                  ...thread,
                  comments: thread.comments.filter(
                    (comment) => comment.id !== context.optimisticCommentId,
                  ),
                }
              : thread,
          ),
        );
      }
      toast.error("Could not add the reply", { description: error.message });
    },
    onSuccess: (comment, { threadId }, context) => {
      queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
        threads?.map((thread) =>
          thread.id === threadId
            ? {
                ...thread,
                comments: reconcileDocumentReply(
                  thread.comments,
                  comment,
                  context.optimisticCommentId,
                ),
              }
            : thread,
        ),
      );
    },
    onSettled: (_data, _error, _payload, context) => {
      context?.release();
      return invalidate();
    },
  });
  const resolve = useMutation({
    mutationKey,
    mutationFn: async ({
      threadId,
      resolved,
    }: {
      threadId: string;
      resolved: boolean;
    }) => {
      const response = await resolveDocumentCommentAction(
        documentId,
        threadId,
        resolved,
        workspaceSlug,
      );
      if (response.error) throw new Error(response.error.message);
    },
    onMutate: async ({ resolved, threadId }) => {
      const release = await prepareDocumentThreadMutation(
        queryClient,
        queryKey,
        threadId,
      );
      const previous = queryClient
        .getQueryData<DocumentCommentThread[]>(queryKey)
        ?.find((thread) => thread.id === threadId);
      const optimistic = {
        resolvedAt: resolved ? new Date().toISOString() : null,
        resolvedBy: resolved ? session?.user.id ?? null : null,
      };
      queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
        threads?.map((thread) =>
          thread.id === threadId ? { ...thread, ...optimistic } : thread,
        ),
      );
      return { threadId, optimistic, previous, release };
    },
    onError: (error, _payload, context) => {
      if (context?.previous) {
        const { threadId, optimistic, previous } = context;
        queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
          threads?.map((thread) =>
            thread.id === threadId &&
            thread.resolvedAt === optimistic.resolvedAt &&
            thread.resolvedBy === optimistic.resolvedBy
              ? {
                  ...thread,
                  resolvedAt: previous.resolvedAt,
                  resolvedBy: previous.resolvedBy,
                }
              : thread,
          ),
        );
      }
      toast.error("Could not update the comment", {
        description: error.message,
      });
    },
    onSettled: (_data, _error, _payload, context) => {
      context?.release();
      return invalidate();
    },
  });
  return {
    create,
    reply,
    resolve,
    isCreating: pendingMutations.some((value) => value?.quote !== undefined),
    isCreatingDraft: (draftTarget: string) =>
      pendingMutations.some((value) => value?.draftTarget === draftTarget),
    isThreadPending: (threadId: string) =>
      isOptimisticThread(threadId) ||
      pendingMutations.some((value) => value?.threadId === threadId),
  };
};

export const useRelatedDocuments = (
  entityType: DocumentRelationType,
  entityId: string,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: documentKeys.related(workspaceSlug, entityType, entityId),
    queryFn: () =>
      getRelatedDocuments(entityType, entityId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(session && entityId),
  });
};

export const useCreateDocument = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: (input?: DocumentCreate) =>
      createDocumentAction(workspaceSlug, input),
    onSuccess: (response) => {
      if (response.error) throw new Error(response.error.message);
      queryClient.invalidateQueries({
        queryKey: documentKeys.lists(workspaceSlug),
      });
    },
    onError: () => toast.error("Could not create the document"),
  });
};

export const useUpdateDocument = (documentId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    scope: { id: `document:${workspaceSlug}:${documentId}:update` },
    mutationFn: (payload: DocumentUpdate) =>
      updateDocumentAction(documentId, payload, workspaceSlug),
    onSuccess: (response) => {
      if (response.error) throw new Error(response.error.message);
      const document = response.data!;
      queryClient.setQueryData<WorkspaceDocument>(
        documentKeys.detail(workspaceSlug, documentId),
        document,
      );
      queryClient.invalidateQueries({
        queryKey: documentKeys.lists(workspaceSlug),
      });
    },
    onError: () => toast.error("Could not save the document"),
  });
};

export const useUpdateDocumentAccess = (documentId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: (payload: DocumentAccessUpdate) =>
      updateDocumentAccessAction(documentId, payload, workspaceSlug),
    onSuccess: (response) => {
      if (response.error) throw new Error(response.error.message);
      queryClient.setQueryData(
        documentKeys.detail(workspaceSlug, documentId),
        response.data,
      );
      toast.success("Document access updated");
      void queryClient.invalidateQueries({
        queryKey: documentKeys.lists(workspaceSlug),
      });
    },
    onError: () => toast.error("Could not update document access"),
  });
};

export const useArchiveDocument = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const restoreDocument = useRestoreArchivedDocument();
  return useMutation({
    mutationFn: async (documentId: string) => {
      const response = await archiveDocumentAction(documentId, workspaceSlug);
      if (response.error) throw new Error(response.error.message);
      return response;
    },
    onSuccess: (_, documentId) => {
      void queryClient.invalidateQueries({
        queryKey: documentKeys.all(workspaceSlug),
      });
      toast.success("Document archived", {
        description: "You can restore it from Archived. Public sharing is off.",
        action: {
          label: "Undo",
          onClick: () => {
            restoreDocument.mutate(documentId);
          },
        },
      });
    },
    onError: () => toast.error("Could not archive the document"),
  });
};

export const useRestoreArchivedDocument = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async (documentId: string) => {
      const response = await restoreArchivedDocumentAction(
        documentId,
        workspaceSlug,
      );
      if (response.error) throw new Error(response.error.message);
      if (!response.data)
        throw new Error("The restored document was not returned");
      return response.data;
    },
    onSuccess: (document) => {
      queryClient.setQueryData(
        documentKeys.detail(workspaceSlug, document.id),
        document,
      );
      void queryClient.invalidateQueries({
        queryKey: documentKeys.all(workspaceSlug),
      });
      toast.success("Document restored", {
        description: "Public sharing stays off. You can enable it from Share.",
      });
    },
    onError: () => toast.error("Could not restore the document"),
  });
};

export const useDuplicateDocument = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: (documentId: string) =>
      duplicateDocumentAction(documentId, workspaceSlug),
    onSuccess: (response) => {
      if (response.error) throw new Error(response.error.message);
      queryClient.invalidateQueries({
        queryKey: documentKeys.lists(workspaceSlug),
      });
      toast.success("Document duplicated");
    },
    onError: () => toast.error("Could not duplicate the document"),
  });
};

export const useDeleteDocument = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: (documentId: string) =>
      deleteDocumentAction(documentId, workspaceSlug),
    onSuccess: (response) => {
      if (response.error) throw new Error(response.error.message);
      queryClient.invalidateQueries({
        queryKey: documentKeys.all(workspaceSlug),
      });
      toast.success("Document deleted");
    },
    onError: () => toast.error("Could not delete the document"),
  });
};

export const useDocumentRelationshipMutations = (documentId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const invalidate = () => {
    queryClient.invalidateQueries({
      queryKey: documentKeys.detail(workspaceSlug, documentId),
    });
    queryClient.invalidateQueries({
      queryKey: documentKeys.all(workspaceSlug),
    });
  };
  const add = useMutation({
    mutationFn: (payload: {
      entityType: DocumentRelationType;
      entityId: string;
    }) => addDocumentRelationshipAction(documentId, payload, workspaceSlug),
    onSuccess: (response) => {
      if (response.error) throw new Error(response.error.message);
      invalidate();
    },
    onError: () => toast.error("Could not relate this work"),
  });
  const remove = useMutation({
    mutationFn: (payload: {
      entityType: DocumentRelationType;
      entityId: string;
    }) =>
      removeDocumentRelationshipAction(
        documentId,
        payload.entityType,
        payload.entityId,
        workspaceSlug,
      ),
    onSuccess: (response) => {
      if (response.error) throw new Error(response.error.message);
      invalidate();
    },
    onError: () => toast.error("Could not remove the relationship"),
  });
  return { add, remove };
};
