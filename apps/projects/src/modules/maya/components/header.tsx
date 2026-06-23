"use client";
import { HistoryIcon, PlusIcon } from "icons";
import { Flex, Button, Text, Tooltip, Box, CircleProgressBar } from "ui";
import { useState } from "react";
import { HistoryDialog } from "@/components/ui/chat/history-dialog";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useSession } from "@/lib/auth/client";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useWorkspacePath } from "@/hooks";
import { useTotalMessages } from "@/modules/ai-chats/hooks/use-total-messages";

type MayaHeaderProps = {
  currentChatId: string;
  handleNewChat: () => void;
  handleChatSelect: (chatId: string) => void;
};

export const Header = ({
  currentChatId,
  handleNewChat,
  handleChatSelect,
}: MayaHeaderProps) => {
  const [isHistoryOpen, setIsHistoryOpen] = useState(false);
  const { data: totalMessages = 0 } = useTotalMessages();
  const { data: session } = useSession();
  const { remaining, getLimit, tier } = useSubscriptionFeatures();
  const { withWorkspace } = useWorkspacePath();
  const isInternalUser = session?.user.isInternal === true;
  const remainingQueries = remaining("maxAiMessages", totalMessages);
  const maxMessages = getLimit("maxAiMessages");
  const usageProgress =
    maxMessages > 0 ? Math.round((totalMessages / maxMessages) * 100) : 0;

  return (
    <>
      <HeaderContainer className="justify-between border-b-0">
        <Flex align="center" gap={2}>
          <MobileMenuButton />
          <Button
            color="tertiary"
            leftIcon={<PlusIcon strokeWidth={2.8} />}
            onClick={handleNewChat}
            variant="naked"
          >
            New chat
          </Button>
        </Flex>
        <Flex align="center" className="gap-2 md:gap-4">
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
              <span className="flex cursor-default items-center gap-2">
                <CircleProgressBar
                  invertColors
                  progress={Math.min(usageProgress, 100)}
                  size={24}
                  strokeWidth={3}
                />
                <Text>
                  {Math.min(totalMessages, getLimit("maxAiMessages"))}/
                  {getLimit("maxAiMessages")}
                </Text>
              </span>
            </Tooltip>
          )}

          <Button
            className="gap-2"
            color="tertiary"
            leftIcon={<HistoryIcon className="h-[1.15rem]" strokeWidth={2.6} />}
            onClick={() => {
              setIsHistoryOpen(true);
            }}
            variant="naked"
          >
            History
          </Button>
        </Flex>
      </HeaderContainer>
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
