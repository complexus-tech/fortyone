"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { toast } from "sonner";
import { useWalkthrough } from "@/components/walkthrough/walkthrough-provider";
import { useMayaChat } from "@/modules/maya/hooks/use-maya-chat";
import { useMayaNavigation } from "@/modules/maya/hooks/use-maya-navigation";
import type { GoogleDriveFileReference } from "@/shared/google-drive/types";

type ChatActions = {
  openChat: (message?: string) => void;
  openChatWithDraft: (draft: string) => void;
  openChatWithGoogleDriveFile: (
    file: Pick<GoogleDriveFileReference, "id" | "mimeType" | "name">,
    draft?: string,
  ) => void;
  closeChat: () => void;
  setIsOpen: (isOpen: boolean) => void;
};

const ChatActionsContext = createContext<ChatActions | null>(null);
const ChatVisibilityContext = createContext<boolean | null>(null);
const ChatRuntimeContext = createContext<ReturnType<typeof useMayaChat> | null>(
  null,
);

export const ChatProvider = ({ children }: { children: ReactNode }) => {
  const {
    currentChatId,
    hasSelectedChat,
    isOnMayaPage,
    routeChatId,
    routeKey,
    updateChatRef,
    clearChatRef,
  } = useMayaNavigation();
  const { completeWalkthroughAction } = useWalkthrough();
  const handleUserMessageSubmitted = useCallback(() => {
    completeWalkthroughAction("maya-message-completed");
  }, [completeWalkthroughAction]);
  // The runtime belongs to the workspace, not the popup or current page.
  // Navigation and closing Maya must not interrupt a response or lose a draft.
  const chat = useMayaChat({
    currentChatId,
    hasSelectedChat,
    onUserMessageSubmitted: handleUserMessageSubmitted,
    updateChatRef,
    clearChatRef,
  });
  const {
    addGoogleDriveFile,
    cancelPendingChatSelection,
    handleChatSelect,
    handleSuggestedPrompt,
    setInput,
  } = chat;
  const lastHandledRouteRef = useRef(routeKey);

  useEffect(() => {
    // A bare Maya URL resumes this workspace's runtime. Only a different,
    // explicit conversation URL requests history; token updates never do.
    if (lastHandledRouteRef.current === routeKey) return;
    lastHandledRouteRef.current = routeKey;
    if (!routeChatId) {
      if (isOnMayaPage) cancelPendingChatSelection();
      return;
    }
    if (routeChatId === currentChatId) {
      cancelPendingChatSelection();
      return;
    }

    void handleChatSelect(routeChatId).catch(() => {
      toast.error("Unable to open that conversation. Please try again.");
    });
  }, [
    cancelPendingChatSelection,
    currentChatId,
    handleChatSelect,
    isOnMayaPage,
    routeChatId,
    routeKey,
  ]);
  const [isOpen, setIsOpen] = useState(false);
  const pendingInitialMessageRef = useRef<string | null>(null);
  const runtimeActionsRef = useRef({
    addGoogleDriveFile,
    handleSuggestedPrompt,
    setInput,
  });
  const isOpenRef = useRef(isOpen);

  useEffect(() => {
    runtimeActionsRef.current = {
      addGoogleDriveFile,
      handleSuggestedPrompt,
      setInput,
    };
  }, [addGoogleDriveFile, handleSuggestedPrompt, setInput]);

  useEffect(() => {
    isOpenRef.current = isOpen;
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen || !pendingInitialMessageRef.current) {
      return;
    }

    const message = pendingInitialMessageRef.current;
    pendingInitialMessageRef.current = null;
    handleSuggestedPrompt(message);
  }, [handleSuggestedPrompt, isOpen]);

  const openChat = useCallback((message?: string) => {
    if (message) {
      if (isOpenRef.current) {
        runtimeActionsRef.current.handleSuggestedPrompt(message);
      } else {
        pendingInitialMessageRef.current = message;
      }
    }
    setIsOpen(true);
  }, []);

  const openChatWithDraft = useCallback((draft: string) => {
    pendingInitialMessageRef.current = null;
    runtimeActionsRef.current.setInput(draft);
    setIsOpen(true);
  }, []);

  const openChatWithGoogleDriveFile = useCallback(
    (
      file: Pick<GoogleDriveFileReference, "id" | "mimeType" | "name">,
      draft?: string,
    ) => {
      pendingInitialMessageRef.current = null;
      runtimeActionsRef.current.addGoogleDriveFile({
        referenceId: file.id,
        mimeType: file.mimeType,
        name: file.name.slice(0, 500),
      });
      if (draft) runtimeActionsRef.current.setInput(draft);
      setIsOpen(true);
    },
    [],
  );

  const closeChat = useCallback(() => {
    setIsOpen(false);
    pendingInitialMessageRef.current = null;
  }, []);

  // Controls elsewhere in the workspace only subscribe to stable actions.
  // A token, draft, or attachment update only notifies runtime consumers.
  const actions = useMemo<ChatActions>(
    () => ({
      closeChat,
      openChat,
      openChatWithDraft,
      openChatWithGoogleDriveFile,
      setIsOpen,
    }),
    [closeChat, openChat, openChatWithDraft, openChatWithGoogleDriveFile],
  );

  return (
    <ChatActionsContext.Provider value={actions}>
      <ChatVisibilityContext.Provider value={isOpen}>
        <ChatRuntimeContext.Provider value={chat}>
          {children}
        </ChatRuntimeContext.Provider>
      </ChatVisibilityContext.Provider>
    </ChatActionsContext.Provider>
  );
};

export const useChatActions = () => {
  const context = useContext(ChatActionsContext);
  if (!context) {
    throw new Error("useChatActions must be used within ChatProvider");
  }
  return context;
};

export const useChatVisibility = () => {
  const context = useContext(ChatVisibilityContext);
  if (context === null) {
    throw new Error("useChatVisibility must be used within ChatProvider");
  }
  return context;
};

export const useChatRuntime = () => {
  const context = useContext(ChatRuntimeContext);
  if (!context) {
    throw new Error("useChatRuntime must be used within ChatProvider");
  }
  return context;
};
