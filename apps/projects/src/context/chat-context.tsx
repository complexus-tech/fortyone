"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import { Chat } from "@/components/ui/chat";

type ChatContextType = {
  openChat: (message?: string) => void;
  closeChat: () => void;
  isOpen: boolean;
};

const ChatContext = createContext<ChatContextType | null>(null);

export const ChatProvider = ({ children }: { children: ReactNode }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [initialMessage, setInitialMessage] = useState<string>("");

  const openChat = (message?: string) => {
    if (message) {
      setInitialMessage(message);
    }
    setIsOpen(true);
  };

  const closeChat = () => {
    setIsOpen(false);
    setInitialMessage("");
  };

  return (
    <ChatContext.Provider value={{ openChat, closeChat, isOpen }}>
      {children}
      <Chat
        initialMessage={initialMessage}
        isOpen={isOpen}
        onMessageSent={() => {
          setInitialMessage("");
        }}
        setIsOpen={setIsOpen}
      />
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
