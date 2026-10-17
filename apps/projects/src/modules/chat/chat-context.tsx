"use client";

import type { ReactNode } from "react";
import { createContext, useContext } from "react";
import type { ChatPerson } from "./types";

type ChatIdentity = {
  userId: string;
  conversationId: string;
  members: ChatPerson[];
  mayaEnabled: boolean;
};

const ChatContext = createContext<ChatIdentity>({
  userId: "",
  conversationId: "",
  members: [],
  mayaEnabled: false,
});

export const ChatIdentityProvider = ({
  children,
  value,
}: {
  children: ReactNode;
  value: ChatIdentity;
}) => <ChatContext.Provider value={value}>{children}</ChatContext.Provider>;

export const useChatIdentity = () => useContext(ChatContext);
