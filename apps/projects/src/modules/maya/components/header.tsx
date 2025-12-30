"use client";
import { HistoryIcon, PlusIcon } from "icons";
import { Flex, Button, Text, Tooltip, Box, CircleProgressBar } from "ui";
import { useState } from "react";
import { HistoryDialog } from "@/components/ui/chat/history-dialog";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useTotalMessages } from "@/modules/ai-chats/hooks/use-total-messages";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";

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
  const { remaining, getLimit, tier } = useSubscriptionFeatures();
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
          {tier !== "enterprise" && (
            <Tooltip
              title={
                <Box className="max-w-xs py-1.5">
                  <Text className="mb-2">
                    You&apos;re remaining with {remainingQueries} of{" "}
                    {getLimit("maxAiMessages")} chat messages. Upgrade for
                    unlimited messages!
                  </Text>
                  <Button
                    color="invert"
                    href="/settings/workspace/billing"
                    fullWidth
                    align="center"
                  >
                    Upgrade plan
                  </Button>
                </Box>
              }
            >
              <span className="flex cursor-default items-center gap-2">
                <CircleProgressBar
                  progress={usageProgress}
                  size={24}
                  strokeWidth={3}
                  invertColors={true}
                />
                <Text>
                  {totalMessages}/{getLimit("maxAiMessages")}
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
