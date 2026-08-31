import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import { useChat } from "@ai-sdk/react";
import { useQueryClient } from "@tanstack/react-query";
import type { FileUIPart } from "ai";
import {
  DefaultChatTransport,
  generateId,
  lastAssistantMessageIsCompleteWithApprovalResponses,
} from "ai";
import { useSession } from "@/lib/auth/client";
import { useSubscription } from "@/lib/hooks/subscriptions/subscription";
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
import { getMayaToolInvalidationKeys } from "../utils/tool-query-invalidation";
import {
  isChatResponseInProgress,
  runWithChatSendGuard,
} from "../utils/chat-send-policy";
import { prepareMayaChatSendRequest } from "../utils/chat-request-payload";
import { canSendMayaMessage } from "../utils/message-limit";
import { mergeRealtimeVoiceMessages } from "../utils/realtime-voice-messages";
import { useMayaRealtimeVoice } from "./use-maya-realtime-voice";

class MayaChatTransport extends DefaultChatTransport<MayaUIMessage> {
  private requestBody: Record<string, unknown> = {};

  constructor() {
    super({ prepareSendMessagesRequest: prepareMayaChatSendRequest });
    this.body = () => this.requestBody;
  }

  setRequestBody(requestBody: Record<string, unknown>) {
    this.requestBody = requestBody;
  }
}

export const useMayaChat = (config: MayaChatConfig) => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const pathname = usePathname();
  const { data: session } = useSession();
  const { data: subscription } = useSubscription();
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
  const isSendingRef = useRef(false);
  const onUserMessageSubmittedRef = useRef(config.onUserMessageSubmitted);

  useEffect(() => {
    onUserMessageSubmittedRef.current = config.onUserMessageSubmitted;
  }, [config.onUserMessageSubmitted]);

  const terminology = {
    stories: getTermDisplay("storyTerm", { variant: "plural" }),
    sprints: getTermDisplay("sprintTerm", { variant: "plural" }),
    objectives: getTermDisplay("objectiveTerm", { variant: "plural" }),
    keyResults: getTermDisplay("keyResultTerm", { variant: "plural" }),
  };
  const requestBody = {
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
    workspace,
    memories,
    terminology,
    totalMessages: {
      current: totalMessages,
      limit: getLimit("maxAiMessages"),
    },
  };
  const transport = useMemo(() => new MayaChatTransport(), []);
  useEffect(() => {
    transport.setRequestBody(requestBody);
  });

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

  const {
    messages,
    status,
    sendMessage,
    stop: handleStop,
    regenerate,
    error,
    setMessages,
    addToolApprovalResponse,
  } = useChat<MayaUIMessage>({
    id: currentChatId,
    transport,
    sendAutomaticallyWhen: lastAssistantMessageIsCompleteWithApprovalResponses,
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

        // Refresh chat list on any text response
        if (part.type === "text") {
          queryClient.invalidateQueries({ queryKey: aiChatKeys.lists() });
          queryClient.invalidateQueries({
            queryKey: aiChatKeys.totalMessages(workspaceSlug),
          });
          return;
        }

        if (
          part.type.startsWith("tool-") &&
          "state" in part &&
          part.state === "output-available"
        ) {
          const toolName = part.type.slice("tool-".length);
          const input = "input" in part ? part.input : undefined;
          for (const queryKey of getMayaToolInvalidationKeys({
            input,
            toolName,
            workspaceSlug,
          })) {
            queryClient.invalidateQueries({ queryKey });
          }
        }
      });
    },
    messages: aiChatMessages,
  });
  const navigateFromVoice = useCallback(
    (path: string) => {
      router.push(withWorkspace(path));
    },
    [router, withWorkspace],
  );
  const setThemeFromVoice = useCallback(
    (requestedTheme: "dark" | "light" | "system" | "toggle") => {
      if (requestedTheme === "toggle") {
        setTheme(resolvedTheme === "dark" ? "light" : "dark");
        return;
      }
      setTheme(requestedTheme);
    },
    [resolvedTheme, setTheme],
  );
  const realtimeVoice = useMayaRealtimeVoice({
    conversationMessages: messages,
    currentPath: pathname,
    navigate: navigateFromVoice,
    setApplicationTheme: setThemeFromVoice,
  });
  const displayMessages = useMemo(
    () => mergeRealtimeVoiceMessages(messages, realtimeVoice.messages),
    [messages, realtimeVoice.messages],
  );

  const handleRegenerate = async (messageId?: string) => {
    await runWithChatSendGuard({
      sendGuard: isSendingRef,
      status,
      task: async () => {
        await regenerate({ messageId });
      },
    });
  };

  const handleSendMessage = async (content: string) => {
    if (!content.trim() && attachments.length === 0) return;
    if (isSendingRef.current) return;
    if (isChatResponseInProgress(status)) return;

    if (content.trim()) {
      onUserMessageSubmittedRef.current?.();
    }

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

    const pendingAttachments = attachments;
    await runWithChatSendGuard({
      sendGuard: isSendingRef,
      status,
      task: async () => {
        // Convert attachments to base64 for AI SDK
        const attachmentData: FileUIPart[] = await Promise.all(
          pendingAttachments.map(async (file) => ({
            type: "file",
            mediaType: file.type,
            filename: file.name,
            url: await fileToBase64(file),
          })),
        );

        setInput((currentInput) =>
          currentInput === content ? "" : currentInput,
        );
        const pendingAttachmentSet = new Set(pendingAttachments);
        setAttachments((currentAttachments) =>
          currentAttachments.filter(
            (attachment) => !pendingAttachmentSet.has(attachment),
          ),
        );

        await sendMessage({
          text: content,
          files: attachmentData,
        });
      },
    });
  };

  const handleSuggestedPrompt = (prompt: string) => {
    void handleSendMessage(prompt);
  };

  const handleSend = () => {
    if (!input.trim() && attachments.length === 0) return;
    void handleSendMessage(input);
  };

  const stop = () => {
    handleStop();
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
    handleStop: stop,
    regenerate: handleRegenerate,
    handleNewChat,
    handleChatSelect,
    handleSuggestedPrompt,
    addToolApprovalResponse,
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
