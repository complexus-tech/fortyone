import { ArrowDown2Icon, HistoryIcon, NewTabIcon, PlusIcon } from "icons";
import { Flex, Button, Text, Tooltip } from "ui";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { HistoryDialog } from "./history-dialog";

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
  return (
    <>
      <Flex align="center" justify="between">
        <Flex align="center" gap={2}>
          <Tooltip title="New chat">
            <Button
              asIcon
              color="tertiary"
              leftIcon={<PlusIcon className="h-[1.4rem]" strokeWidth={2.8} />}
              onClick={handleNewChat}
              variant="naked"
            >
              <span className="sr-only">New chat</span>
            </Button>
          </Tooltip>
          <Tooltip title="History">
            <Button
              asIcon
              color="tertiary"
              leftIcon={
                <HistoryIcon className="h-[1.4rem]" strokeWidth={2.8} />
              }
              onClick={() => {
                setIsHistoryOpen(true);
              }}
              variant="naked"
            >
              <span className="sr-only">History</span>
            </Button>
          </Tooltip>
        </Flex>
        <Text className="antialiased" fontWeight="semibold">
          Maya is your AI assistant
        </Text>
        <Flex align="center" gap={3}>
          <Tooltip title="Open on new page">
            <Button
              asIcon
              color="tertiary"
              leftIcon={<NewTabIcon className="h-[1.4rem]" strokeWidth={2.6} />}
              onClick={() => {
                router.push(`/maya?chatRef=${currentChatId}`);
                setIsOpen(false);
              }}
              variant="naked"
            >
              <span className="sr-only">Open on new page</span>
            </Button>
          </Tooltip>
          <Tooltip title="Close">
            <Button
              asIcon
              color="tertiary"
              leftIcon={<ArrowDown2Icon className="h-6" strokeWidth={2.8} />}
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
