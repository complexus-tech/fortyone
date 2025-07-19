import { useRef, useState, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useChat } from "@ai-sdk/react";
import { useQueryClient } from "@tanstack/react-query";
import { generateId } from "ai";
import { useSession } from "next-auth/react";
import {
  notificationKeys,
  sprintKeys,
  statusKeys,
  teamKeys,
} from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import { objectiveKeys } from "@/modules/objectives/constants";
import { useSubscription } from "@/lib/hooks/subscriptions/subscription";
import { useTeams } from "@/modules/teams/hooks/teams";
import { fileToBase64 } from "@/lib/utils/files";
import { useAiChatMessages } from "@/modules/ai-chats/hooks/use-ai-chat-messages";
import { getAiChatMessages } from "@/modules/ai-chats/queries/get-ai-chat-messages";
import { aiChatKeys } from "@/modules/ai-chats/constants";
import type { MayaChatConfig } from "../types";
import { useMayaNavigation } from "./use-maya-navigation";

export const useMayaChat = (config: MayaChatConfig) => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const pathname = usePathname();
  const { data: session } = useSession();
  const { data: subscription } = useSubscription();
  const { data: teams = [] } = useTeams();
  const { updateChatRef, clearChatRef } = useMayaNavigation();

  const { resolvedTheme, theme, setTheme } = useTheme();
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isObjectiveOpen, setIsObjectiveOpen] = useState(false);
  const [isSprintOpen, setIsSprintOpen] = useState(false);
  const [attachments, setAttachments] = useState<File[]>([]);
  const idRef = useRef(config.currentChatId);
  const { data: aiChatMessages = [] } = useAiChatMessages(idRef.current);

  const handleNewChat = () => {
    const newChatId = generateId();
    idRef.current = newChatId;
    setMessages([]);
    setInput("");
    setAttachments([]);
    clearChatRef();
  };

  const handleChatSelect = async (chatId: string) => {
    idRef.current = chatId;
    // Fetch messages for the new chat ID directly
    const newMessages = await queryClient.fetchQuery({
      queryKey: aiChatKeys.messages(chatId),
      queryFn: () => getAiChatMessages(session!, chatId),
    });
    setMessages(newMessages);
    setInput("");
    setAttachments([]);
    updateChatRef(chatId);
  };

  const {
    messages,
    input,
    status,
    setInput,
    append,
    stop: handleStop,
    reload,
    error,
    setMessages,
  } = useChat({
    body: {
      currentPath: pathname,
      currentTheme: theme,
      resolvedTheme,
      subscription: {
        tier: subscription?.tier,
        billingInterval: subscription?.billingInterval,
        billingEndsAt: subscription?.billingEndsAt,
        status: subscription?.status,
      },
      teams,
      id: idRef.current,
    },
    sendExtraMessageFields: true,
    onFinish: (message) => {
      message.parts?.forEach((part) => {
        if (part.type === "tool-invocation") {
          if (part.toolInvocation.toolName === "navigation") {
            if (part.toolInvocation.state === "result") {
              const result = part.toolInvocation.result;
              if (result.route) {
                router.push(result.route as string);
              }
            }
          } else if (part.toolInvocation.toolName === "theme") {
            if (part.toolInvocation.state === "result") {
              const requestedTheme = part.toolInvocation.result.theme as string;
              if (requestedTheme === "toggle") {
                const newTheme = resolvedTheme === "dark" ? "light" : "dark";
                setTheme(newTheme);
              } else {
                setTheme(requestedTheme);
              }
            }
          } else if (part.toolInvocation.toolName === "quickCreate") {
            if (part.toolInvocation.state === "result") {
              const action = part.toolInvocation.result.action as string;
              switch (action) {
                case "story":
                  setIsStoryOpen(true);
                  break;
                case "objective":
                  setIsObjectiveOpen(true);
                  break;
                case "sprint":
                  setIsSprintOpen(true);
                  break;
              }
            }
          } else if (part.toolInvocation.toolName === "teams") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: teamKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "notifications") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: notificationKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "statuses") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: statusKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "stories") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: storyKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "sprints") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: sprintKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "objectives") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: objectiveKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "objectiveStatuses") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: objectiveKeys.statuses(),
              });
            }
          }
        } else if (part.type === "text") {
          queryClient.invalidateQueries({
            queryKey: aiChatKeys.lists(),
          });
        }
      });
    },
    initialMessages: aiChatMessages,
  });

  // Sync messages when aiChatMessages changes (when selecting a different chat)
  useEffect(() => {
    if (aiChatMessages.length > 0) {
      setMessages(aiChatMessages);
    } else if (aiChatMessages.length === 0) {
      // Clear messages when no messages are available
      setMessages([]);
    }
  }, [aiChatMessages, setMessages]);

  // Clear messages when starting a new chat (no chatRef)
  useEffect(() => {
    if (config.isNewChat) {
      setMessages([]);
      setInput("");
      setAttachments([]);
    }
  }, [config.isNewChat, setMessages, setInput, setAttachments]);

  const isLoading = status === "submitted";

  const handleSendMessage = async (content: string) => {
    if (!content.trim() && attachments.length === 0) return;

    // Convert attachments to base64 for AI SDK
    const attachmentData = await Promise.all(
      attachments.map(async (file) => ({
        name: file.name,
        contentType: file.type,
        url: await fileToBase64(file),
      })),
    );

    append({
      role: "user",
      content:
        content ||
        `Analyze the attached file${attachmentData.length > 1 ? "s" : ""}.`,
      experimental_attachments:
        attachmentData.length > 0 ? attachmentData : undefined,
    });

    setInput("");
    setAttachments([]);
  };

  const handleSuggestedPrompt = (prompt: string) => {
    handleSendMessage(prompt);
  };

  const handleSend = () => {
    if (!input.trim() && attachments.length === 0) return;
    handleSendMessage(input);
  };

  return {
    // Chat state
    messages,
    input,
    status,
    isLoading,
    error,
    attachments,
    currentChatId: idRef.current,

    // Chat actions
    setInput,
    handleSend,
    handleStop,
    reload,
    handleNewChat,
    handleChatSelect,
    handleSuggestedPrompt,
    setAttachments,

    // Dialog states
    isStoryOpen,
    setIsStoryOpen,
    isObjectiveOpen,
    setIsObjectiveOpen,
    isSprintOpen,
    setIsSprintOpen,
  };
};
