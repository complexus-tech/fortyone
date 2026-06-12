import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { generateId } from "ai";
import { useState } from "react";

export const useMayaNavigation = () => {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const chatRef = searchParams.get("chatRef");
  const [draftChatId, setDraftChatId] = useState(() => generateId());
  const currentChatId = chatRef || draftChatId;

  const getInitialChatId = (): string => {
    return currentChatId;
  };

  const isNewChat = (): boolean => {
    return !chatRef;
  };

  const updateChatRef = (chatId: string) => {
    const params = new URLSearchParams(searchParams);
    params.set("chatRef", chatId);
    router.replace(`${pathname}?${params.toString()}`);
  };

  const clearChatRef = (nextChatId = generateId()) => {
    setDraftChatId(nextChatId);
    const params = new URLSearchParams(searchParams);
    params.delete("chatRef");
    router.replace(pathname);
  };

  return {
    chatRef,
    getInitialChatId,
    isNewChat,
    updateChatRef,
    clearChatRef,
  };
};
