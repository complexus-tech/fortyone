import { useMemo, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import { useChat } from "@ai-sdk/react";
import { useQueryClient } from "@tanstack/react-query";
import type { FileUIPart } from "ai";
import { generateId } from "ai";
import { useSession } from "@/lib/auth/client";
import {
  githubKeys,
  integrationRequestKeys,
  labelKeys,
  notificationKeys,
  teamKeys,
  calendarKeys,
} from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import { objectiveKeys } from "@/modules/objectives/constants";
import { useSubscription } from "@/lib/hooks/subscriptions/subscription";
import { useTeams } from "@/modules/teams/hooks/teams";
import { fileToBase64 } from "@/lib/utils/files";
import { useAiChatMessages } from "@/modules/ai-chats/hooks/use-ai-chat-messages";
import { getAiChatMessages } from "@/modules/ai-chats/queries/get-ai-chat-messages";
import { aiChatKeys } from "@/modules/ai-chats/constants";
import { useProfile } from "@/lib/hooks/profile";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useMemories } from "@/modules/ai-chats/hooks/use-memory";
import { useTotalMessages } from "@/modules/ai-chats/hooks/use-total-messages";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import type { MayaChatConfig } from "../types";
import { canSendMayaMessage } from "../utils/message-limit";
import { mergeRealtimeVoiceMessages } from "../utils/realtime-voice-messages";
import { useMayaRealtimeVoice } from "./use-maya-realtime-voice";

export const useMayaChat = (config: MayaChatConfig) => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const pathname = usePathname();
  const { data: session } = useSession();
  const { data: subscription } = useSubscription();
  const { data: teams = [] } = useTeams();
  const { data: profile } = useProfile();
  const { data: memories = [] } = useMemories();
  const { data: totalMessages = 0 } = useTotalMessages();
  const { getLimit, displayTier } = useSubscriptionFeatures();
  const isInternalUser = session?.user.isInternal === true;
  const { workspace } = useCurrentWorkspace();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const { resolvedTheme, theme, setTheme } = useTheme();
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isObjectiveOpen, setIsObjectiveOpen] = useState(false);
  const [isSprintOpen, setIsSprintOpen] = useState(false);
  const [attachments, setAttachments] = useState<File[]>([]);
  const { getTermDisplay } = useTerminology();
  const currentChatId = config.currentChatId;
  const { data: aiChatMessages = [] } = useAiChatMessages(currentChatId);
  const [input, setInput] = useState("");

  const terminology = {
    stories: getTermDisplay("storyTerm", { variant: "plural" }),
    sprints: getTermDisplay("sprintTerm", { variant: "plural" }),
    objectives: getTermDisplay("objectiveTerm", { variant: "plural" }),
    keyResults: getTermDisplay("keyResultTerm", { variant: "plural" }),
  };

  const handleNewChat = () => {
    const newChatId = generateId();
    realtimeVoice.disconnect();
    realtimeVoice.clearMessages();
    config.clearChatRef(newChatId);
    setMessages([]);
    setInput("");
    setAttachments([]);
  };

  const handleChatSelect = async (chatId: string) => {
    realtimeVoice.disconnect();
    realtimeVoice.clearMessages();
    // Fetch messages for the new chat ID directly
    const newMessages = await queryClient.fetchQuery({
      queryKey: aiChatKeys.messages(chatId),
      queryFn: () =>
        getAiChatMessages({ session: session!, workspaceSlug }, chatId),
    });
    setMessages(newMessages);
    setInput("");
    setAttachments([]);
    config.updateChatRef(chatId);
  };

  // Map tool types to the query keys they should invalidate
  const toolInvalidationMap: Partial<Record<string, () => readonly unknown[]>> =
    {
      // Teams
      "tool-createTeamTool": () => teamKeys.all(workspaceSlug),
      "tool-updateTeam": () => teamKeys.all(workspaceSlug),
      "tool-joinTeam": () => teamKeys.all(workspaceSlug),
      "tool-leaveTeam": () => teamKeys.all(workspaceSlug),
      "tool-deleteTeam": () => teamKeys.all(workspaceSlug),
      // Stories
      "tool-createStory": () => storyKeys.all(workspaceSlug),
      "tool-updateStory": () => storyKeys.all(workspaceSlug),
      "tool-deleteStory": () => storyKeys.all(workspaceSlug),
      "tool-bulkUpdateStories": () => storyKeys.all(workspaceSlug),
      "tool-bulkDeleteStories": () => storyKeys.all(workspaceSlug),
      "tool-bulkCreateStories": () => storyKeys.all(workspaceSlug),
      "tool-assignStoriesToUser": () => storyKeys.all(workspaceSlug),
      "tool-duplicateStory": () => storyKeys.all(workspaceSlug),
      "tool-restoreStory": () => storyKeys.all(workspaceSlug),
      "tool-addStoryAssociation": () => storyKeys.all(workspaceSlug),
      "tool-removeStoryAssociation": () => storyKeys.all(workspaceSlug),
      "tool-comments": () => storyKeys.all(workspaceSlug),
      "tool-storyLabels": () => storyKeys.all(workspaceSlug),
      "tool-links": () => storyKeys.all(workspaceSlug),
      "tool-labels": () => labelKeys.all(workspaceSlug),
      // Integration requests
      "tool-updateIntegrationRequestTool": () =>
        integrationRequestKeys.all(workspaceSlug),
      "tool-acceptIntegrationRequestTool": () =>
        integrationRequestKeys.all(workspaceSlug),
      "tool-declineIntegrationRequestTool": () =>
        integrationRequestKeys.all(workspaceSlug),
      "tool-acceptAllIntegrationRequestsTool": () =>
        integrationRequestKeys.all(workspaceSlug),
      "tool-declineAllIntegrationRequestsTool": () =>
        integrationRequestKeys.all(workspaceSlug),
      "tool-postRequestGitHubCommentTool": () =>
        integrationRequestKeys.all(workspaceSlug),
      // Objectives & Key Results
      "tool-createKeyResultTool": () => objectiveKeys.all(workspaceSlug),
      "tool-updateKeyResultTool": () => objectiveKeys.all(workspaceSlug),
      "tool-deleteKeyResultTool": () => objectiveKeys.all(workspaceSlug),
      "tool-createObjectiveTool": () => objectiveKeys.all(workspaceSlug),
      "tool-updateObjectiveTool": () => objectiveKeys.all(workspaceSlug),
      "tool-deleteObjectiveTool": () => objectiveKeys.all(workspaceSlug),
      // Notifications
      "tool-notifications": () => notificationKeys.all(workspaceSlug),
      // GitHub
      "tool-resyncGitHubRepositoriesTool": () =>
        githubKeys.integration(workspaceSlug),
      "tool-createGitHubIssueSyncLinkTool": () =>
        githubKeys.integration(workspaceSlug),
      "tool-deleteGitHubIssueSyncLinkTool": () =>
        githubKeys.integration(workspaceSlug),
      "tool-updateGitHubWorkspaceSettingsTool": () =>
        githubKeys.integration(workspaceSlug),
      "tool-updateGitHubTeamSettingsTool": () =>
        githubKeys.integration(workspaceSlug),
      "tool-postStoryGitHubCommentTool": () => storyKeys.all(workspaceSlug),
      "tool-deleteStoryGitHubLinkTool": () => storyKeys.all(workspaceSlug),
      // Memory
      "tool-deleteMemory": () => aiChatKeys.memories(),
      "tool-updateMemory": () => aiChatKeys.memories(),
      "tool-createMemory": () => aiChatKeys.memories(),
      // Objective statuses
      "tool-objectiveStatuses": () => objectiveKeys.statuses(workspaceSlug),
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
    id: currentChatId,
    onFinish: ({ message }) => {
      message.parts.forEach((part) => {
        // Handle side effects for navigation and theme
        if (part.type === "tool-navigation") {
          if (part.state === "output-available" && part.output.route) {
            router.push(part.output.route);
          }
          return;
        }

        if (part.type === "tool-theme") {
          if (part.state === "output-available") {
            const requested = part.output.theme;
            if (requested === "toggle") {
              setTheme(resolvedTheme === "dark" ? "light" : "dark");
            } else {
              setTheme(requested);
            }
          }
          return;
        }

        if (
          part.type === "tool-mayaWorkPlanTool" &&
          part.state === "output-available"
        ) {
          queryClient.invalidateQueries({
            queryKey: calendarKeys.all(workspaceSlug),
          });
          queryClient.invalidateQueries({
            queryKey: storyKeys.all(workspaceSlug),
          });
          return;
        }

        // Refresh chat list on any text response
        if (part.type === "text") {
          queryClient.invalidateQueries({ queryKey: aiChatKeys.lists() });
          queryClient.invalidateQueries({
            queryKey: aiChatKeys.totalMessages(),
          });
          return;
        }

        // Invalidate queries for tool completions via the map
        const getQueryKey = toolInvalidationMap[part.type];
        if (
          getQueryKey &&
          "state" in part &&
          part.state === "output-available"
        ) {
          queryClient.invalidateQueries({ queryKey: getQueryKey() });
        }
      });
    },
    messages: aiChatMessages,
  });
  const realtimeVoice = useMayaRealtimeVoice({
    conversationMessages: messages,
    currentPath: pathname,
  });
  const displayMessages = useMemo(
    () => mergeRealtimeVoiceMessages(messages, realtimeVoice.messages),
    [messages, realtimeVoice.messages],
  );

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
        memories,
        totalMessages: {
          current: totalMessages,
          limit: getLimit("maxAiMessages"),
        },
        workspace,
        terminology,
        id: currentChatId,
      },
    });
  };

  const handleSendMessage = async (content: string) => {
    if (!content.trim() && attachments.length === 0) return;

    const limit = getLimit("maxAiMessages");
    if (
      !canSendMayaMessage({
        isInternalUser,
        limit,
        totalMessages,
      })
    ) {
      toast.error("Message limit reached", {
        description: `You have reached your monthly limit of ${limit} messages for the ${displayTier} plan. Please upgrade to continue chatting.`,
        action: {
          label: "Upgrade",
          onClick: () => {
            router.push(withWorkspace("/settings/workspace/billing"));
          },
        },
        duration: 4000,
      });
      return;
    }

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
          memories,
          terminology,
          totalMessages: {
            current: totalMessages,
            limit: getLimit("maxAiMessages"),
          },
          id: currentChatId,
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
    displayMessages,
    input,
    status,
    error,
    attachments,
    currentChatId,
    realtimeVoice,

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
