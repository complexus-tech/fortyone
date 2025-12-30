import { CloseIcon, HistoryIcon, MinusIcon, NewTabIcon, PlusIcon } from "icons";
import { Flex, Button, Text, Tooltip, Badge, Box, CircleProgressBar } from "ui";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useAiChats } from "@/modules/ai-chats/hooks/use-ai-chats";
import { HistoryDialog } from "./history-dialog";
import { useTotalMessages } from "@/modules/ai-chats/hooks/use-total-messages";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";

export const ChatHeader = ({
  currentChatId,
  setIsOpen,
  handleNewChat,
  handleChatSelect,
}: {
  currentChatId: string;
  setIsOpen: (isOpen: boolean) => void;
  handleNewChat: () => void;
  handleChatSelect: (chatId: string) => void;
}) => {
  const router = useRouter();
  const [isHistoryOpen, setIsHistoryOpen] = useState(false);
  const { data: chats = [] } = useAiChats();
  const { data: totalMessages = 0 } = useTotalMessages();
  const { remaining, tier, getLimit } = useSubscriptionFeatures();
  const remainingQueries = remaining("maxAiMessages", totalMessages);
  const maxMessages = getLimit("maxAiMessages");
  const usageProgress =
    maxMessages > 0 ? Math.round((totalMessages / maxMessages) * 100) : 0;
  return (
    <>
      <Flex align="center" justify="between">
        <Flex align="center" gap={2}>
          <Tooltip title="New chat">
            <Button
              asIcon
              color="tertiary"
              leftIcon={
                <PlusIcon
                  className="text-dark dark:text-gray-200"
                  strokeWidth={2.8}
                />
              }
              onClick={handleNewChat}
              variant="naked"
            >
              <span className="sr-only">New chat</span>
            </Button>
          </Tooltip>
          <Tooltip title="Open on new page">
            <Button
              asIcon
              color="tertiary"
              leftIcon={
                <NewTabIcon
                  className="text-dark dark:text-gray-200"
                  strokeWidth={2.6}
                />
              }
              onClick={() => {
                router.push(`/maya?chatRef=${currentChatId}`);
                setIsOpen(false);
              }}
              variant="naked"
            >
              <span className="sr-only">Open on new page</span>
            </Button>
          </Tooltip>
        </Flex>
        <Text
          className="flex items-center gap-2 text-xl antialiased"
          fontWeight="semibold"
        >
          Maya <Badge className="rounded-lg">Beta</Badge>
        </Text>
        <Flex align="center" gap={3}>
          {remainingQueries <= 20 && (
            <Box className="hidden w-[60px] shrink-0 items-center justify-center md:flex">
              <CircleProgressBar
                progress={usageProgress}
                size={32}
                strokeWidth={3}
              >
                <Text className="text-xs font-medium">{remainingQueries}</Text>
              </CircleProgressBar>
            </Box>
          )}

          {chats.length > 0 && (
            <Tooltip title="History">
              <Button
                asIcon
                color="tertiary"
                leftIcon={
                  <HistoryIcon
                    className="text-dark dark:text-gray-200"
                    strokeWidth={2.8}
                  />
                }
                onClick={() => {
                  setIsHistoryOpen(true);
                }}
                variant="naked"
              >
                <span className="sr-only">History</span>
              </Button>
            </Tooltip>
          )}

          <Tooltip title="Close">
            <Button
              asIcon
              color="tertiary"
              leftIcon={
                <CloseIcon
                  className="text-dark dark:text-gray-200"
                  strokeWidth={2.8}
                />
              }
              onClick={() => {
                setIsOpen(false);
              }}
              variant="naked"
            >
              <span className="sr-only">Close</span>
            </Button>
          </Tooltip>
        </Flex>
      </Flex>
      <HistoryDialog
        currentChatId={currentChatId}
        handleChatSelect={handleChatSelect}
        handleNewChat={handleNewChat}
        isOpen={isHistoryOpen}
        setIsOpen={setIsHistoryOpen}
      />
    </>
  );
};
