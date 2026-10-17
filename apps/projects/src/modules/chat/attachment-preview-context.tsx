"use client";

import type { ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
} from "react";
import { useQuery } from "@tanstack/react-query";
import { ReplyIcon } from "icons";
import { Button } from "ui";
import { AttachmentPreviewDialog } from "@/components/ui/attachment-preview";
import { getPreviewMediaKind } from "@/components/ui/attachment-preview/types";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getMessage } from "./api";
import { getChatAttachmentUrl, getChatPdfPreviewUrl } from "./attachment-url";
import { chatKeys } from "./constants";
import type { ChatMessage, ChatResource } from "./types";

type PreviewSession = {
  messageId: string;
  conversationId: string;
  selectedId: string;
  message?: ChatMessage;
  trigger: HTMLButtonElement;
  surface: HTMLElement | null;
};

type PreviewActions = {
  openMessage: (
    message: ChatMessage,
    attachmentId: string,
    trigger: HTMLButtonElement,
  ) => void;
  openResource: (resource: ChatResource, trigger: HTMLButtonElement) => void;
};

const PreviewContext = createContext<PreviewActions | null>(null);

const findSurface = (trigger: HTMLButtonElement) =>
  trigger.closest<HTMLElement>(
    'section[aria-label$=" conversation"], aside[aria-label="Thread"]',
  );

export const ChatAttachmentPreviewProvider = ({
  children,
  conversationName,
  onViewMessage,
}: {
  children: ReactNode;
  conversationName?: string;
  onViewMessage: (messageId: string) => void;
}) => {
  const { workspaceSlug } = useWorkspacePath();
  const [session, setSession] = useState<PreviewSession | null>(null);
  const returningToMessage = useRef(false);
  const messageId = session?.messageId ?? "";
  const message = useQuery({
    queryKey: [...chatKeys.all(workspaceSlug), "message", messageId],
    queryFn: ({ signal }) => getMessage({ workspaceSlug }, messageId, signal),
    enabled: Boolean(session),
    initialData: session?.message,
    staleTime: 0,
    gcTime: 60_000,
    retry: false,
  });
  const openMessage = useCallback<PreviewActions["openMessage"]>(
    (source, selectedId, trigger) => {
      if (
        source.deletedAt ||
        !source.attachments.some(
          (item) =>
            item.id === selectedId && getPreviewMediaKind(item) !== "audio",
        )
      )
        return;
      returningToMessage.current = false;
      setSession({
        messageId: source.id,
        conversationId: source.conversationId,
        selectedId,
        message: source,
        trigger,
        surface: findSurface(trigger),
      });
    },
    [],
  );
  const openResource = useCallback<PreviewActions["openResource"]>(
    (resource, trigger) => {
      if (resource.kind !== "file" || getPreviewMediaKind(resource) === "audio")
        return;
      returningToMessage.current = false;
      setSession({
        messageId: resource.messageId,
        conversationId: resource.conversationId,
        selectedId: resource.attachmentId,
        trigger,
        surface: findSurface(trigger),
      });
    },
    [],
  );
  const actions = useMemo(
    () => ({ openMessage, openResource }),
    [openMessage, openResource],
  );
  const source = message.data;
  const unavailable = Boolean(
    source &&
      (source.deletedAt || source.conversationId !== session?.conversationId),
  );
  const attachments = useMemo(() => {
    if (!session || unavailable || message.isError) return [];
    return (source?.attachments ?? [])
      .filter((item) => getPreviewMediaKind(item) !== "audio")
      .map((item) => ({
        ...item,
        url:
          getPreviewMediaKind(item) === "pdf"
            ? getChatPdfPreviewUrl(
                workspaceSlug,
                session.conversationId,
                item.id,
              )
            : getChatAttachmentUrl(
                workspaceSlug,
                session.conversationId,
                item.id,
              ),
      }));
  }, [
    message.isError,
    session,
    source?.attachments,
    unavailable,
    workspaceSlug,
  ]);
  const removed = Boolean(
    source &&
      !unavailable &&
      !attachments.some((item) => item.id === session?.selectedId),
  );
  let error: string | undefined;
  if (message.isError)
    error =
      "This file could not be loaded. It may have been removed, or your access may have changed.";
  else if (unavailable || removed)
    error = "This attachment is no longer available in this message.";

  return (
    <PreviewContext.Provider value={actions}>
      {children}
      {session ? (
        <AttachmentPreviewDialog
          attachments={attachments}
          contextLabel={
            conversationName
              ? `Shared in ${conversationName}`
              : "Message attachments"
          }
          error={error}
          footer={
            <Button
              color="tertiary"
              leftIcon={<ReplyIcon className="size-4" />}
              onClick={() => {
                returningToMessage.current = true;
                setSession(null);
                onViewMessage(session.messageId);
              }}
              size="sm"
              variant="naked"
            >
              View in conversation
            </Button>
          }
          loading={message.isPending}
          onClose={() => {
            setSession(null);
          }}
          onCloseAutoFocus={(event) => {
            event.preventDefault();
            if (returningToMessage.current) return;
            if (session.trigger.isConnected) {
              session.trigger.focus();
              return;
            }
            const fallback = session.surface?.querySelector<HTMLElement>(
              'button[aria-label="Close thread"], h2[tabindex="-1"], button[aria-label="Search this conversation"]',
            );
            fallback?.focus();
          }}
          onRetry={() => {
            void message.refetch();
          }}
          onSelect={(selectedId) => {
            setSession((current) =>
              current ? { ...current, selectedId } : current,
            );
          }}
          open
          selectedId={session.selectedId}
        />
      ) : null}
    </PreviewContext.Provider>
  );
};

export const useChatAttachmentPreview = () => {
  const context = useContext(PreviewContext);
  if (!context)
    throw new Error("Chat attachments require ChatAttachmentPreviewProvider");
  return context;
};
