"use client";

import { useEffect, useRef } from "react";
import { usePathname } from "next/navigation";
import { useMediaQuery } from "@/hooks";
import { useChatContext } from "@/context/chat-context";
import { ChatContent } from "./content";

export const ChatRail = () => {
  const pathname = usePathname();
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { closeChat, isOpen } = useChatContext();
  const popupRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen || !isDesktop || pathname.includes("maya")) {
      return;
    }

    const focusFrame = window.requestAnimationFrame(() => {
      popupRef.current
        ?.querySelector<HTMLTextAreaElement>('[aria-label="Chat message"]')
        ?.focus();
    });
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        closeChat();
      }
    };
    window.addEventListener("keydown", handleEscape);

    return () => {
      window.cancelAnimationFrame(focusFrame);
      window.removeEventListener("keydown", handleEscape);
    };
  }, [closeChat, isDesktop, isOpen, pathname]);

  if (!isOpen || !isDesktop || pathname.includes("maya")) {
    return null;
  }

  return (
    <div
      aria-label="Chat with Maya"
      aria-modal="false"
      className="border-border dark:bg-surface animate-in fade-in zoom-in-95 slide-in-from-bottom-2 fixed right-[22px] bottom-6 z-50 h-[min(760px,calc(100dvh-64px))] w-[min(420px,calc(100vw-36px))] origin-bottom-right overflow-hidden rounded-lg border bg-white shadow-none backdrop-blur-2xl duration-200"
      ref={popupRef}
      role="dialog"
    >
      <ChatContent isPopup />
    </div>
  );
};
