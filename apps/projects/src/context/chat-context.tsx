"use client";

import { generateId } from "ai";
import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { useMayaChat } from "@/modules/maya";

type ChatContextType = {
  chat: ReturnType<typeof useMayaChat>;
  openChat: (message?: string) => void;
  closeChat: () => void;
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
};

const ChatContext = createContext<ChatContextType | null>(null);

export const ChatProvider = ({ children }: { children: ReactNode }) => {
  const [chatId] = useState(() => generateId());
  const chat = useMayaChat({
    currentChatId: chatId,
  });
  const [isOpen, setIsOpen] = useState(false);
  const pendingInitialMessageRef = useRef<string | null>(null);

  useEffect(() => {
    if (!isOpen || !pendingInitialMessageRef.current) {
      return;
    }

    const message = pendingInitialMessageRef.current;
    pendingInitialMessageRef.current = null;
    chat.handleSuggestedPrompt(message);
  }, [chat.handleSuggestedPrompt, isOpen]);

  const openChat = (message?: string) => {
    if (message) {
      if (isOpen) {
        chat.handleSuggestedPrompt(message);
      } else {
        pendingInitialMessageRef.current = message;
      }
    }
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
