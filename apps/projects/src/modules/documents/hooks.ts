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
  deleteDocumentAction,
  duplicateDocumentAction,
  removeDocumentRelationshipAction,
  updateDocumentAccessAction,
  updateDocumentAction,
} from "./actions";
import { getDocument, getDocuments, getRelatedDocuments } from "./queries";
import type {
  DocumentAccessUpdate,
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
