import { useRef, useState, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useChat } from "@ai-sdk/react";
import { useQueryClient } from "@tanstack/react-query";
import type { FileUIPart } from "ai";
import { generateId } from "ai";
import { useSession } from "next-auth/react";
import { notificationKeys, sprintKeys, teamKeys } from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import { objectiveKeys } from "@/modules/objectives/constants";
import { useSubscription } from "@/lib/hooks/subscriptions/subscription";
import { useTeams } from "@/modules/teams/hooks/teams";
import { fileToBase64 } from "@/lib/utils/files";
import { useAiChatMessages } from "@/modules/ai-chats/hooks/use-ai-chat-messages";
import { getAiChatMessages } from "@/modules/ai-chats/queries/get-ai-chat-messages";
import { aiChatKeys } from "@/modules/ai-chats/constants";
import { useProfile } from "@/lib/hooks/profile";
import { useTerminology } from "@/hooks";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import type { MayaChatConfig } from "../types";
import { useMayaNavigation } from "./use-maya-navigation";

export const useMayaChat = (config: MayaChatConfig) => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const pathname = usePathname();
  const { data: session } = useSession();
  const { data: subscription } = useSubscription();
  const { data: teams = [] } = useTeams();
  const { data: profile } = useProfile();
  const { workspace } = useCurrentWorkspace();
  const { updateChatRef, clearChatRef } = useMayaNavigation();
  const { resolvedTheme, theme, setTheme } = useTheme();
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isObjectiveOpen, setIsObjectiveOpen] = useState(false);
  const [isSprintOpen, setIsSprintOpen] = useState(false);
  const [attachments, setAttachments] = useState<File[]>([]);
  const { getTermDisplay } = useTerminology();
  const idRef = useRef(config.currentChatId);
  const { data: aiChatMessages = [] } = useAiChatMessages(idRef.current);
  const [input, setInput] = useState("");

  const terminology = {
    stories: getTermDisplay("storyTerm", { variant: "plural" }),
    sprints: getTermDisplay("sprintTerm", { variant: "plural" }),
    objectives: getTermDisplay("objectiveTerm", { variant: "plural" }),
    keyResults: getTermDisplay("keyResultTerm", { variant: "plural" }),
  };

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
    status,
    sendMessage,
    stop: handleStop,
    regenerate,
    error,
    setMessages,
  } = useChat<MayaUIMessage>({
    onFinish: ({ message }) => {
      message.parts.forEach((part) => {
        if (part.type === "tool-navigation") {
          if (part.state === "output-available") {
            const route = part.output.route;
            if (route) {
              router.push(route);
            }
          }
        } else if (part.type === "tool-theme") {
          if (part.state === "output-available") {
            const requestedTheme = part.output.theme;
            if (requestedTheme === "toggle") {
              const newTheme = resolvedTheme === "dark" ? "light" : "dark";
              setTheme(newTheme);
            } else {
              setTheme(requestedTheme);
            }
          }
        } else if (part.type === "text") {
          queryClient.invalidateQueries({
            queryKey: aiChatKeys.lists(),
          });
        } else if (
          part.type === "tool-createTeamTool" ||
          part.type === "tool-updateTeam" ||
          part.type === "tool-joinTeam" ||
          part.type === "tool-leaveTeam" ||
          part.type === "tool-deleteTeam"
        ) {
          if (part.state === "output-available") {
            queryClient.invalidateQueries({
              queryKey: teamKeys.all,
            });
          }
        } else if (
          part.type === "tool-createStory" ||
          part.type === "tool-updateStory" ||
          part.type === "tool-deleteStory" ||
          part.type === "tool-bulkUpdateStories" ||
          part.type === "tool-bulkDeleteStories" ||
          part.type === "tool-bulkCreateStories" ||
          part.type === "tool-assignStoriesToUser" ||
          part.type === "tool-duplicateStory" ||
          part.type === "tool-restoreStory"
        ) {
          if (part.state === "output-available") {
            queryClient.invalidateQueries({
              queryKey: storyKeys.all,
            });
          }
        } else if (part.type === "tool-createSprint") {
          if (part.state === "output-available") {
            queryClient.invalidateQueries({
              queryKey: sprintKeys.all,
            });
          }
        } else if (
          part.type === "tool-createKeyResultTool" ||
          part.type === "tool-updateKeyResultTool" ||
          part.type === "tool-deleteKeyResultTool" ||
          part.type === "tool-createObjectiveTool" ||
          part.type === "tool-updateObjectiveTool" ||
          part.type === "tool-deleteObjectiveTool"
        ) {
          if (part.state === "output-available") {
            queryClient.invalidateQueries({
              queryKey: objectiveKeys.all,
            });
          }
        } else if (part.type === "tool-notifications") {
          if (part.state === "output-available") {
            queryClient.invalidateQueries({
              queryKey: notificationKeys.all,
            });
          }
        } else if (part.type === "tool-objectiveStatuses") {
          if (part.state === "output-available") {
            queryClient.invalidateQueries({
              queryKey: objectiveKeys.statuses(),
            });
          }
        }
      });
    },
    messages: aiChatMessages,
  });

  const handleRegenerate = (messageId?: string) => {
    regenerate({
      messageId,
      body: {
        currentPath: pathname,
        currentTheme: theme,
        resolvedTheme,
        subscription: {
          tier: subscription?.tier,
          billingInterval: subscription?.billingInterval,
          billingEndsAt: subscription?.billingEndsAt,
          status: subscription?.status,
          username: profile?.username,
        },
        teams,
        workspace,
        terminology,
        id: idRef.current,
      },
    });
  };

  // Clear messages when starting a new chat (no chatRef)
  useEffect(() => {
    if (config.isNewChat) {
      setMessages([]);
      setInput("");
      setAttachments([]);
    }
  }, [config.isNewChat, setMessages, setInput, setAttachments]);

  const handleSendMessage = async (content: string) => {
    if (!content.trim() && attachments.length === 0) return;

    // Convert attachments to base64 for AI SDK
    const attachmentData: FileUIPart[] = await Promise.all(
      attachments.map(async (file) => ({
        type: "file",
        mediaType: file.type,
        name: file.name,
        url: await fileToBase64(file),
      })),
    );

    sendMessage(
      {
        text: content,
        files: attachmentData,
      },
      {
        body: {
          currentPath: pathname,
          currentTheme: theme,
          resolvedTheme,
          subscription: {
            tier: subscription?.tier,
            billingInterval: subscription?.billingInterval,
            billingEndsAt: subscription?.billingEndsAt,
            status: subscription?.status,
            username: profile?.username,
          },
          teams,
          workspace,
          terminology,
          id: idRef.current,
        },
      },
    );

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
    error,
    attachments,
    currentChatId: idRef.current,

    // Chat actions
    setInput,
    handleSend,
    handleStop,
    regenerate: handleRegenerate,
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
