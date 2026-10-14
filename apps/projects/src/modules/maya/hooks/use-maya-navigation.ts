import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { generateId } from "ai";
import { useCallback, useEffect, useRef, useState } from "react";

type ActiveChat = {
  id: string;
  hasSelectedChat: boolean;
};

export const useMayaNavigation = () => {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const isOnMayaPage = pathname.endsWith("/maya");
  const routeChatId = isOnMayaPage ? searchParams.get("chatRef") : null;
  const [activeChat, setActiveChat] = useState<ActiveChat>(() => ({
    id: routeChatId || generateId(),
    hasSelectedChat: Boolean(routeChatId),
  }));
  const navigationRef = useRef({
    isOnMayaPage,
    pathname,
    router,
    searchParams,
  });

  useEffect(() => {
    navigationRef.current = { isOnMayaPage, pathname, router, searchParams };
  }, [isOnMayaPage, pathname, router, searchParams]);

  const updateChatRef = useCallback((chatId: string) => {
    setActiveChat({ hasSelectedChat: true, id: chatId });
    const navigation = navigationRef.current;
    if (
      !navigation.isOnMayaPage ||
      navigation.searchParams.get("chatRef") === chatId
    ) {
      return;
    }

    const params = new URLSearchParams(navigation.searchParams);
    params.set("chatRef", chatId);
    navigation.router.replace(`${navigation.pathname}?${params.toString()}`, {
      scroll: false,
    });
  }, []);

  const clearChatRef = useCallback((nextChatId = generateId()) => {
    setActiveChat({ hasSelectedChat: false, id: nextChatId });
    const navigation = navigationRef.current;
    if (!navigation.isOnMayaPage || !navigation.searchParams.has("chatRef")) {
      return;
    }

    const params = new URLSearchParams(navigation.searchParams);
    params.delete("chatRef");
    const queryString = params.toString();
    navigation.router.replace(
      queryString
        ? `${navigation.pathname}?${queryString}`
        : navigation.pathname,
      {
        scroll: false,
      },
    );
  }, []);

  return {
    currentChatId: activeChat.id,
    hasSelectedChat: activeChat.hasSelectedChat,
    isOnMayaPage,
    routeChatId,
    routeKey: `${pathname}?chatRef=${routeChatId ?? ""}`,
    updateChatRef,
    clearChatRef,
  };
};
