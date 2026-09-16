import { useCallback, useEffect, useMemo, useSyncExternalStore } from "react";
import { AppState } from "react-native";
import { useFocusEffect } from "expo-router";
import { Directory, File, FileMode, Paths } from "expo-file-system";
import { randomUUID } from "expo-crypto";
import { useAuthStore } from "@/store/auth";
import {
  createMayaAttachmentQueue,
  getMayaAttachmentMediaType,
  MAYA_ATTACHMENT_MEDIA_TYPES,
  validateMayaAttachmentSizes,
  type MayaAttachment,
} from "../lib/attachments";
import { matchesMayaSession } from "../lib/session-scope";

const createNativeAttachmentIO = (isCurrent: () => boolean) => {
  const ownedFiles = new Map<string, File>();
  const directories = new Set<Directory>();
  const release = (attachments: readonly MayaAttachment[]) => {
    for (const attachment of attachments) {
      const file = ownedFiles.get(attachment.id);
      ownedFiles.delete(attachment.id);
      // Only files created by this adapter can be removed, never picker originals.
      try {
        if (file?.exists) file.delete();
      } catch {
        // This remains an OS-evictable cache file if storage is temporarily unavailable.
      }
    }
    for (const directory of directories) {
      try {
        if (directory.exists && directory.list().length === 0)
          directory.delete();
        if (!directory.exists) directories.delete(directory);
      } catch {
        // Keep the directory for another cleanup attempt; no file content is logged.
      }
    }
  };
  return {
    release,
    pick: async (): Promise<MayaAttachment[]> => {
      const result = await File.pickFileAsync({
        multipleFiles: true,
        mimeTypes: [...MAYA_ATTACHMENT_MEDIA_TYPES],
      });
      if (result.canceled) return [];
      if (!isCurrent())
        throw new Error("The conversation changed. Select files again.");
      // Native stat validation precedes copying or allocating base64 data.
      validateMayaAttachmentSizes(
        result.result.map((file) => ({ size: file.size })),
      );
      const directory = new Directory(
        Paths.cache,
        `maya-attachments-${randomUUID()}`,
      );
      directory.create();
      directories.add(directory);
      const attachments: MayaAttachment[] = [];
      try {
        for (const source of result.result) {
          const id = randomUUID();
          const file = new File(directory, id);
          ownedFiles.set(id, file);
          const attachment: MayaAttachment = {
            id,
            uri: file.uri,
            name:
              source.name.replace(/[\u0000-\u001f\u007f]/g, "").slice(0, 255) ||
              "Attachment",
            mediaType: "application/pdf",
            size: source.size,
          };
          attachments.push(attachment);
          await source.copy(file);
          if (!isCurrent())
            throw new Error("The conversation changed. Select files again.");
          const handle = file.open(FileMode.ReadOnly);
          try {
            attachment.size = handle.size ?? 0;
            validateMayaAttachmentSizes(attachments);
            attachment.mediaType = getMayaAttachmentMediaType(
              handle.readBytes(16),
            );
          } finally {
            handle.close();
          }
        }
        return attachments;
      } catch (error) {
        release(attachments);
        throw error instanceof Error &&
          /conversation changed|Attach|attachment/i.test(error.message)
          ? error
          : new Error(
              "A selected file could not be read. Choose it again from Files.",
            );
      }
    },
    read: async (attachment: MayaAttachment) => {
      const file = ownedFiles.get(attachment.id);
      if (!file?.exists || file.size !== attachment.size)
        throw new Error(
          "An attachment is no longer available. Select it again.",
        );
      const size = file.size;
      validateMayaAttachmentSizes([{ size }]);
      const base64 = await file.base64();
      if (!isCurrent() || !file.exists || file.size !== size)
        throw new Error("The attachment changed. Select it again.");
      return { size, base64 };
    },
  };
};

export const useMayaAttachments = (scopeKey: string) => {
  const identity = useAuthStore((state) =>
    state.isAuthenticated && !state.isLoading
      ? `${state.userId}:${state.workspace}:${state.sessionEpoch}`
      : null,
  );
  const queue = useMemo(() => {
    const state = useAuthStore.getState();
    const scope = {
      userId: state.userId ?? "",
      workspace: state.workspace ?? "",
      sessionEpoch: state.sessionEpoch,
    };
    const isCurrent = () =>
      Boolean(scopeKey) &&
      identity !== null &&
      matchesMayaSession(scope, useAuthStore.getState());
    return createMayaAttachmentQueue({
      isCurrent,
      ...createNativeAttachmentIO(isCurrent),
    });
    // The conversation key gives every chat a separate queue and cache ownership.
  }, [scopeKey, identity]);
  const snapshot = useSyncExternalStore(
    queue.subscribe,
    queue.getSnapshot,
    queue.getSnapshot,
  );
  useFocusEffect(
    useCallback(() => {
      queue.focus();
      return () => queue.blur();
    }, [queue]),
  );
  useEffect(() => {
    const appState = AppState.addEventListener("change", (state) => {
      if (state === "background") queue.background();
    });
    return () => {
      appState.remove();
      queue.blur();
    };
  }, [queue]);
  return {
    ...snapshot,
    pick: queue.pick,
    remove: queue.remove,
    clear: queue.clear,
    prepare: queue.prepare,
  };
};
