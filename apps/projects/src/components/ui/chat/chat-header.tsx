import { CloseIcon, HistoryIcon, NewTabIcon, PlusIcon } from "icons";
import {
  Avatar,
  Flex,
  Button,
  Text,
  Tooltip,
  Box,
  CircleProgressBar,
} from "ui";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "@/lib/auth/client";
import { useProfile } from "@/lib/hooks/profile";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useWorkspacePath } from "@/hooks";
import { useAiChats } from "@/modules/ai-chats/hooks/use-ai-chats";
import { useTotalMessages } from "@/modules/ai-chats/hooks/use-total-messages";
import { HistoryDialog } from "./history-dialog";

export const ChatHeader = ({
  currentChatId,
  setIsOpen,
  handleNewChat,
  handleChatSelect,
  isPopup = false,
  hasMessages = false,
}: {
  currentChatId: string;
  setIsOpen: (isOpen: boolean) => void;
  handleNewChat: () => void;
  handleChatSelect: (chatId: string) => void;
  isPopup?: boolean;
  hasMessages?: boolean;
}) => {
  const router = useRouter();
  const [isHistoryOpen, setIsHistoryOpen] = useState(false);
  const { data: chats = [] } = useAiChats();
  const { data: totalMessages = 0 } = useTotalMessages();
  const { data: session } = useSession();
  const { data: profile } = useProfile();
  const { remaining, getLimit, tier } = useSubscriptionFeatures();
  const { withWorkspace } = useWorkspacePath();
  const isInternalUser = session?.user.isInternal === true;
  const remainingQueries = remaining("maxAiMessages", totalMessages);
  const maxMessages = getLimit("maxAiMessages");
  const usageProgress =
    maxMessages > 0 ? Math.round((totalMessages / maxMessages) * 100) : 0;

  return (
    <>
      <Flex align="center" className="w-full" justify="between">
        <Flex align="center" className="min-w-0 flex-1" gap={3}>
          {isPopup ? (
            <Avatar
              className="h-10 w-10"
              name={
                profile?.fullName || profile?.username || session?.user.name
              }
              rounded="full"
              size="md"
              src={profile?.avatarUrl || session?.user.image}
            />
          ) : null}
          {isPopup ? (
            <Box className="min-w-0">
              <Text className="text-lg leading-6" fontWeight="medium">
                Maya
              </Text>
              <Text className="text-base leading-6" color="muted">
                AI project assistant
              </Text>
            </Box>
          ) : (
            <Tooltip title="New chat">
              <Button
                asIcon
                color="tertiary"
                leftIcon={
                  <PlusIcon className="text-foreground/70" strokeWidth={2.8} />
                }
                onClick={handleNewChat}
                size="sm"
                variant="naked"
              >
                <span className="sr-only">New chat</span>
              </Button>
            </Tooltip>
          )}
          {!isPopup ? (
            <Tooltip title="Open on new page">
              <Button
                asIcon
                color="tertiary"
                leftIcon={
                  <NewTabIcon
                    className="text-foreground/70"
                    strokeWidth={2.6}
                  />
                }
                onClick={() => {
                  router.push(withWorkspace(`/maya?chatRef=${currentChatId}`));
                  setIsOpen(false);
                }}
                size="sm"
                variant="naked"
              >
                <span className="sr-only">Open on new page</span>
              </Button>
            </Tooltip>
          ) : null}
        </Flex>
        {!isPopup ? (
          <Text
            className="shrink-0 px-3 text-base antialiased"
            fontWeight="semibold"
          >
            Chat with Maya
          </Text>
        ) : null}
        <Flex align="center" className="min-w-0 flex-1 justify-end" gap={2}>
          {tier !== "enterprise" && !isInternalUser && (
            <Tooltip
              title={
                <Box className="max-w-xs py-1.5">
                  <Text className="mb-2">
                    You&apos;re remaining with {remainingQueries} of{" "}
                    {getLimit("maxAiMessages")} chat messages. Upgrade to send
                    more messages!
                  </Text>
                  <Button
                    align="center"
                    color="invert"
                    fullWidth
                    href={withWorkspace("/settings/workspace/billing")}
                  >
                    Upgrade plan
                  </Button>
                </Box>
              }
            >
              <span className="flex cursor-default">
                <CircleProgressBar
                  invertColors
                  progress={usageProgress}
                  size={remainingQueries >= 100 ? 18 : 22}
                  strokeWidth={3}
                >
                  {remainingQueries < 100 ? (
                    <Text className="max-w-[2ch] truncate text-xs font-semibold">
                      {remainingQueries}
                    </Text>
                  ) : null}
                </CircleProgressBar>
              </span>
            </Tooltip>
          )}
          {chats.length > 0 && (
            <Tooltip title="History">
              <Button
                asIcon
                color="tertiary"
                leftIcon={
                  <HistoryIcon
                    className="text-foreground/70"
                    strokeWidth={2.8}
                  />
                }
                onClick={() => {
                  setIsHistoryOpen(true);
                }}
                size="sm"
                variant="naked"
              >
                <span className="sr-only">History</span>
              </Button>
            </Tooltip>
          )}
          {isPopup && hasMessages ? (
            <Tooltip title="New chat">
              <Button
                asIcon
                color="tertiary"
                leftIcon={
                  <PlusIcon className="text-foreground/70" strokeWidth={2.8} />
                }
                onClick={handleNewChat}
                size="sm"
                variant="naked"
              >
                <span className="sr-only">New chat</span>
              </Button>
            </Tooltip>
          ) : null}

          <Tooltip title="Close">
            <Button
              asIcon
              color="tertiary"
              leftIcon={
                <CloseIcon className="text-foreground/70" strokeWidth={2.8} />
              }
              onClick={() => {
                setIsOpen(false);
              }}
              size="sm"
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
