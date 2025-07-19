import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { generateId } from "ai";

export const useMayaNavigation = () => {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const chatRef = searchParams.get("chatRef");

  const getInitialChatId = (): string => {
    return chatRef || generateId();
  };

  const isNewChat = (): boolean => {
    return !chatRef;
  };

  const updateChatRef = (chatId: string) => {
    const params = new URLSearchParams(searchParams);
    params.set("chatRef", chatId);
    router.replace(`${pathname}?${params.toString()}`);
  };

  const clearChatRef = () => {
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
