"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { documentKeys } from "@/shared/documents/keys";
import {
  attachGoogleDriveFilesAction,
  createGoogleDriveConnectSessionAction,
  createGoogleDriveFileAction,
  createGoogleDrivePickerSessionAction,
  deleteGoogleDriveFileAction,
  disconnectGoogleDriveAction,
  importGoogleDriveFileAction,
  refreshGoogleDriveFileAction,
} from "./actions";
import { getGoogleDriveFiles, getGoogleDriveIntegration } from "./queries";
import type {
  GoogleDriveFileReference,
  GoogleDriveFileType,
  GoogleDriveImportVisibility,
  GoogleDrivePickerFile,
  GoogleDriveTarget,
} from "./types";

export const googleDriveKeys = {
  all: (workspaceSlug: string) => ["google-drive", workspaceSlug] as const,
  integration: (workspaceSlug: string) =>
    [...googleDriveKeys.all(workspaceSlug), "integration"] as const,
  files: (workspaceSlug: string, target: GoogleDriveTarget) =>
    [
      ...googleDriveKeys.all(workspaceSlug),
      "files",
      target.type,
      target.id,
    ] as const,
};

const requireData = <T>(
  response: {
    data?: T | null;
    error?: { code?: string; hint?: string; message?: string } | null;
  },
  fallback: string,
) => {
  if (response.error?.message || !response.data) {
    throw Object.assign(new Error(response.error?.message ?? fallback), {
      code: response.error?.code,
      hint: response.error?.hint,
    });
  }
  return response.data;
};

export const useGoogleDriveIntegration = ({ enabled = true } = {}) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    enabled: enabled && Boolean(session && workspaceSlug),
    queryKey: googleDriveKeys.integration(workspaceSlug),
    queryFn: () =>
      getGoogleDriveIntegration({ session: session!, workspaceSlug }),
  });
};

export const useGoogleDriveFiles = (target: GoogleDriveTarget) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    enabled: Boolean(session && workspaceSlug && target.id),
    queryKey: googleDriveKeys.files(workspaceSlug, target),
    queryFn: () =>
      getGoogleDriveFiles({ session: session!, workspaceSlug }, target),
    staleTime: 60_000,
  });
};

export const useCreateGoogleDriveConnectSession = () => {
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async (returnUrl?: string) =>
      requireData(
        await createGoogleDriveConnectSessionAction(workspaceSlug, returnUrl),
        "Google did not return an authorization URL.",
      ),
    onSuccess: ({ authorizationUrl }) => {
      window.location.assign(authorizationUrl);
    },
    onError: (error) =>
      toast.error("Google Drive", { description: error.message }),
  });
};

export const useDisconnectGoogleDrive = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async () => {
      const response = await disconnectGoogleDriveAction(workspaceSlug);
      if (response.error?.message) throw new Error(response.error.message);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: googleDriveKeys.all(workspaceSlug),
      });
      toast.success("Google Drive disconnected");
    },
    onError: (error) =>
      toast.error("Google Drive", { description: error.message }),
  });
};

export const useCreateGoogleDrivePickerSession = () => {
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async () =>
      requireData(
        await createGoogleDrivePickerSessionAction(workspaceSlug),
        "Google Picker could not be opened.",
      ),
  });
};

export const useAttachGoogleDriveFiles = (
  target: GoogleDriveTarget,
  {
    notifyOnError = true,
    notifyOnSuccess = true,
  }: { notifyOnError?: boolean; notifyOnSuccess?: boolean } = {},
) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async (files: GoogleDrivePickerFile[]) =>
      requireData(
        await attachGoogleDriveFilesAction(workspaceSlug, target, files),
        "The selected Google files could not be attached.",
      ),
    onSuccess: (files, selectedFiles) => {
      queryClient.setQueryData<GoogleDriveFileReference[]>(
        googleDriveKeys.files(workspaceSlug, target),
        (current) => {
          const byId = new Map((current ?? []).map((file) => [file.id, file]));
          files.forEach((file) => byId.set(file.id, file));
          return [...byId.values()];
        },
      );
      void queryClient.invalidateQueries({
        queryKey: googleDriveKeys.files(workspaceSlug, target),
      });
      if (notifyOnSuccess) {
        toast.success(
          selectedFiles.length === 1
            ? "Google file attached"
            : "Google files attached",
        );
      }
    },
    onError: (error) => {
      if (notifyOnError) {
        toast.error("Google Drive", { description: error.message });
      }
    },
  });
};

export const useDeleteGoogleDriveFile = (target: GoogleDriveTarget) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async (referenceId: string) => {
      const response = await deleteGoogleDriveFileAction(
        workspaceSlug,
        referenceId,
      );
      if (response.error?.message) throw new Error(response.error.message);
      return referenceId;
    },
    onSuccess: (referenceId) => {
      queryClient.setQueryData<GoogleDriveFileReference[]>(
        googleDriveKeys.files(workspaceSlug, target),
        (files) => files?.filter((file) => file.id !== referenceId) ?? [],
      );
    },
    onError: (error) =>
      toast.error("Google Drive", { description: error.message }),
  });
};

export const useRefreshGoogleDriveFile = (target: GoogleDriveTarget) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async (referenceId: string) =>
      requireData(
        await refreshGoogleDriveFileAction(workspaceSlug, referenceId),
        "The Google file preview could not be refreshed.",
      ),
    onSuccess: (refreshedFile) => {
      queryClient.setQueryData<GoogleDriveFileReference[]>(
        googleDriveKeys.files(workspaceSlug, target),
        (files) =>
          files?.map((file) =>
            file.id === refreshedFile.id ? refreshedFile : file,
          ) ?? [refreshedFile],
      );
    },
    onError: (error) =>
      toast.error("Google Drive", { description: error.message }),
  });
};

export const useCreateGoogleDriveFile = (target: GoogleDriveTarget) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async ({
      fileType,
      idempotencyKey,
      title,
    }: {
      fileType: GoogleDriveFileType;
      idempotencyKey: string;
      title: string;
    }) =>
      requireData(
        await createGoogleDriveFileAction(
          workspaceSlug,
          target,
          fileType,
          title,
          idempotencyKey,
        ),
        "The Google file could not be created.",
      ),
    onSuccess: (file) => {
      queryClient.setQueryData<GoogleDriveFileReference[]>(
        googleDriveKeys.files(workspaceSlug, target),
        (files) => {
          const byId = new Map((files ?? []).map((item) => [item.id, item]));
          byId.set(file.id, file);
          return [...byId.values()];
        },
      );
      toast.success(
        file.mimeType.includes("spreadsheet")
          ? "Google Sheet created"
          : "Google Doc created",
      );
    },
    onError: (error) =>
      toast.error("Google Drive", { description: error.message }),
  });
};

export const useImportGoogleDriveFile = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async ({
      idempotencyKey,
      referenceId,
      visibility,
    }: {
      idempotencyKey: string;
      referenceId: string;
      visibility: GoogleDriveImportVisibility;
    }) =>
      requireData(
        await importGoogleDriveFileAction(
          workspaceSlug,
          referenceId,
          visibility,
          idempotencyKey,
        ),
        "The Google Doc could not be imported.",
      ),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: documentKeys.lists(workspaceSlug),
      });
    },
    onError: (error) =>
      toast.error("Google Drive", { description: error.message }),
  });
};
