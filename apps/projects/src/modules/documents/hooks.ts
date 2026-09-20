"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { documentKeys } from "@/shared/documents/keys";
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
  return useQuery({
    queryKey: documentKeys.comments(workspaceSlug, documentId),
    queryFn: () =>
      getDocumentComments(documentId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(session && documentId),
  });
};

export const useDocumentCommentMutations = (documentId: string) => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = documentKeys.comments(workspaceSlug, documentId);
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
    mutationFn: async (payload: {
      body: string;
      quote: string;
      anchorStart: number;
      anchorEnd: number;
    }) => {
      const response = await createDocumentCommentAction(
        documentId,
        payload,
        workspaceSlug,
      );
      if (response.error || !response.data)
        throw new Error(
          response.error?.message ?? "The comment was not returned",
        );
      return response.data;
    },
    onMutate: async (payload) => {
      await queryClient.cancelQueries({ queryKey });
      const previous =
        queryClient.getQueryData<DocumentCommentThread[]>(queryKey);
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
      return { optimisticThreadId: thread.id, previous };
    },
    onError: (error, _payload, context) => {
      queryClient.setQueryData(queryKey, context?.previous);
      toast.error("Could not add the comment", {
        description: error.message,
      });
    },
    onSuccess: (thread, _payload, context) => {
      queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
        threads?.map((cached) =>
          cached.id === context.optimisticThreadId ? thread : cached,
        ),
      );
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey }),
  });
  const reply = useMutation({
    mutationFn: async ({
      threadId,
      body,
    }: {
      threadId: string;
      body: string;
    }) => {
      const response = await replyToDocumentCommentAction(
        documentId,
        threadId,
        body,
        workspaceSlug,
      );
      if (response.error || !response.data)
        throw new Error(
          response.error?.message ?? "The reply was not returned",
        );
      return response.data;
    },
    onMutate: async ({ body, threadId }) => {
      await queryClient.cancelQueries({ queryKey });
      const previous =
        queryClient.getQueryData<DocumentCommentThread[]>(queryKey);
      const comment = optimisticComment(body);
      queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
        threads?.map((thread) =>
          thread.id === threadId
            ? { ...thread, comments: [...thread.comments, comment] }
            : thread,
        ),
      );
      return { optimisticCommentId: comment.id, previous, threadId };
    },
    onError: (error, _payload, context) => {
      queryClient.setQueryData(queryKey, context?.previous);
      toast.error("Could not add the reply", { description: error.message });
    },
    onSuccess: (comment, _payload, context) => {
      queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
        threads?.map((thread) =>
          thread.id === context.threadId
            ? {
                ...thread,
                comments: thread.comments.map((cached) =>
                  cached.id === context.optimisticCommentId ? comment : cached,
                ),
              }
            : thread,
        ),
      );
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey }),
  });
  const resolve = useMutation({
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
      await queryClient.cancelQueries({ queryKey });
      const previous =
        queryClient.getQueryData<DocumentCommentThread[]>(queryKey);
      queryClient.setQueryData<DocumentCommentThread[]>(queryKey, (threads) =>
        threads?.map((thread) =>
          thread.id === threadId
            ? {
                ...thread,
                resolvedAt: resolved ? new Date().toISOString() : null,
                resolvedBy: resolved ? session?.user.id ?? null : null,
              }
            : thread,
        ),
      );
      return { previous };
    },
    onError: (error, _payload, context) => {
      queryClient.setQueryData(queryKey, context?.previous);
      toast.error("Could not update the comment", {
        description: error.message,
      });
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey }),
  });
  return { create, reply, resolve };
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
    },
    onError: () => toast.error("Could not update document access"),
  });
};

export const useArchiveDocument = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: (documentId: string) =>
      archiveDocumentAction(documentId, workspaceSlug),
    onSuccess: (response) => {
      if (response.error) throw new Error(response.error.message);
      queryClient.invalidateQueries({
        queryKey: documentKeys.all(workspaceSlug),
      });
      toast.success("Document archived");
    },
    onError: () => toast.error("Could not archive the document"),
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
