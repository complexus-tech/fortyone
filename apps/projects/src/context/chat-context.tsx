"use client";

import { generateId } from "ai";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { useWalkthrough } from "@/components/walkthrough/walkthrough-provider";
import { useMayaChat } from "@/modules/maya";
import type { GoogleDriveFileReference } from "@/shared/google-drive/types";

type ChatContextType = {
  chat: ReturnType<typeof useMayaChat>;
  openChat: (message?: string) => void;
  openChatWithDraft: (draft: string) => void;
  openChatWithGoogleDriveFile: (
    file: Pick<GoogleDriveFileReference, "id" | "mimeType" | "name">,
    draft?: string,
  ) => void;
  closeChat: () => void;
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
};

const ChatContext = createContext<ChatContextType | null>(null);

type ActiveChat = {
  id: string;
  hasSelectedChat: boolean;
};

export const ChatProvider = ({ children }: { children: ReactNode }) => {
  const [activeChat, setActiveChat] = useState<ActiveChat>(() => ({
    id: generateId(),
    hasSelectedChat: false,
  }));
  const { completeWalkthroughAction } = useWalkthrough();
  const handleUserMessageSubmitted = useCallback(() => {
    completeWalkthroughAction("maya-message-completed");
  }, [completeWalkthroughAction]);
  const chat = useMayaChat({
    currentChatId: activeChat.id,
    hasSelectedChat: activeChat.hasSelectedChat,
    onUserMessageSubmitted: handleUserMessageSubmitted,
    updateChatRef: (chatId) => {
      setActiveChat({ hasSelectedChat: true, id: chatId });
    },
    clearChatRef: (nextChatId = generateId()) => {
      setActiveChat({ hasSelectedChat: false, id: nextChatId });
    },
  });
  const { handleSuggestedPrompt } = chat;
  const [isOpen, setIsOpen] = useState(false);
  const pendingInitialMessageRef = useRef<string | null>(null);
  const handleSuggestedPromptRef = useRef(handleSuggestedPrompt);
  const isOpenRef = useRef(isOpen);

  useEffect(() => {
    handleSuggestedPromptRef.current = handleSuggestedPrompt;
  }, [handleSuggestedPrompt]);

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
        handleSuggestedPromptRef.current(message);
      } else {
        pendingInitialMessageRef.current = message;
      }
    }
    setIsOpen(true);
  }, []);

  const openChatWithDraft = (draft: string) => {
    pendingInitialMessageRef.current = null;
    chat.setInput(draft);
    setIsOpen(true);
  };
  const { addGoogleDriveFile, setInput } = chat;

  const openChatWithGoogleDriveFile = (
    file: Pick<GoogleDriveFileReference, "id" | "mimeType" | "name">,
    draft?: string,
  ) => {
    pendingInitialMessageRef.current = null;
    addGoogleDriveFile({
      referenceId: file.id,
      mimeType: file.mimeType,
      name: file.name.slice(0, 500),
    });
    if (draft) setInput(draft);
    setIsOpen(true);
  };

  const closeChat = () => {
    setIsOpen(false);
    pendingInitialMessageRef.current = null;
  };

  return (
    <ChatContext.Provider
      value={{
        chat,
        closeChat,
        isOpen,
        openChat,
        openChatWithDraft,
        openChatWithGoogleDriveFile,
        setIsOpen,
      }}
    >
      {children}
    </ChatContext.Provider>
  );
};

export const useChatContext = () => {
  const context = useContext(ChatContext);
  if (!context) {
    throw new Error("useChatContext must be used within ChatProvider");
  }
  return context;
};
