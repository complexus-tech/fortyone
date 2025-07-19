import { ArrowDown2Icon, HistoryIcon, NewTabIcon, PlusIcon } from "icons";
import { Flex, Button, Text, Tooltip } from "ui";
import { useAiChats } from "@/modules/ai-chats/hooks/use-ai-chats";

export const ChatHeader = ({
  currentChatId,
  setIsOpen,
  handleNewChat,
}: {
  currentChatId: string;
  setIsOpen: (isOpen: boolean) => void;
  handleNewChat: () => void;
}) => {
  const { data: aiChats = [] } = useAiChats();
  return (
    <Flex align="center" justify="between">
      <Flex align="center" gap={2}>
        <Tooltip title="New chat">
          <Button
            asIcon
            color="tertiary"
            leftIcon={<PlusIcon className="h-[1.4rem]" strokeWidth={2.8} />}
            onClick={handleNewChat}
            size="sm"
            variant="naked"
          >
            <span className="sr-only">New chat</span>
          </Button>
        </Tooltip>
        <Tooltip title="History">
          <Button
            asIcon
            color="tertiary"
            disabled={aiChats.length === 0}
            leftIcon={<HistoryIcon className="h-[1.4rem]" strokeWidth={2.8} />}
            size="sm"
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
        <Tooltip title="New tab">
          <Button
            asIcon
            color="tertiary"
            href={`/maya?chatRef=${currentChatId}`}
            leftIcon={<NewTabIcon className="h-[1.4rem]" strokeWidth={2.6} />}
            size="sm"
            variant="naked"
          >
            <span className="sr-only">New tab</span>
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
            size="sm"
            variant="naked"
          >
            <span className="sr-only">Close</span>
          </Button>
        </Tooltip>
      </Flex>
    </Flex>
  );
};
