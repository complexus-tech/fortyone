import {
  AiIcon,
  CloseIcon,
  HistoryIcon,
  MaximizeIcon,
  MinimizeIcon,
  PlusIcon,
} from "icons";
import { Flex, Button, Text, Tooltip } from "ui";
import { useMediaQuery } from "@/hooks";
import { useAiChats } from "@/modules/ai-chats/hooks/use-ai-chats";

const BackIcon = () => {
  return (
    <svg
      className="h-6 w-auto -rotate-90 text-gray dark:text-gray-300"
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M18 9.47326L16.5858 10.8813L13.0006 7.31184L13.0006 20.5H11.0006L11.0006 7.3114L7.41422 10.8814L6 9.47338L12.0003 3.5L18 9.47326Z"
        fill="currentColor"
      />
    </svg>
  );
};

export const ChatHeader = ({
  setIsOpen,
  isFullScreen,
  setIsFullScreen,
  handleNewChat,
  isHistoryOpen,
  setIsHistoryOpen,
}: {
  setIsOpen: (isOpen: boolean) => void;
  isFullScreen: boolean;
  setIsFullScreen: (isFullScreen: boolean) => void;
  handleNewChat: () => void;
  isHistoryOpen: boolean;
  setIsHistoryOpen: (isHistoryOpen: boolean) => void;
}) => {
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { data: aiChats = [] } = useAiChats();
  return (
    <Flex align="center" justify="between">
      {isHistoryOpen ? (
        <Flex align="center" gap={2}>
          <Tooltip title="Back">
            <Button
              asIcon
              color="tertiary"
              leftIcon={<BackIcon />}
              onClick={() => {
                setIsHistoryOpen(false);
              }}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">Back</span>
            </Button>
          </Tooltip>
          <Text>All chats</Text>
        </Flex>
      ) : (
        <Flex align="center" gap={2}>
          <Tooltip title="History">
            <Button
              asIcon
              color="tertiary"
              disabled={aiChats.length === 0}
              leftIcon={<HistoryIcon className="h-[1.3rem]" />}
              onClick={() => {
                setIsHistoryOpen(true);
              }}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">History</span>
            </Button>
          </Tooltip>
          <Tooltip title="New chat">
            <Button
              asIcon
              color="tertiary"
              leftIcon={<PlusIcon className="h-[1.3rem]" strokeWidth={2.8} />}
              onClick={handleNewChat}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">New chat</span>
            </Button>
          </Tooltip>
        </Flex>
      )}

      {!isHistoryOpen && (
        <Flex align="center" gap={1}>
          <AiIcon className="h-[1.3rem]" />
          <Text>Maya is your AI assistant</Text>
        </Flex>
      )}
      <Flex align="center" gap={2}>
        {isDesktop ? (
          <Tooltip title={isFullScreen ? "Minimize" : "Maximize"}>
            <Button
              asIcon
              color="tertiary"
              leftIcon={
                isFullScreen ? (
                  <MinimizeIcon className="h-[1.3rem]" strokeWidth={2.8} />
                ) : (
                  <MaximizeIcon className="h-[1.3rem]" strokeWidth={2.8} />
                )
              }
              onClick={() => {
                setIsFullScreen(!isFullScreen);
              }}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">Maximize</span>
            </Button>
          </Tooltip>
        ) : null}
        <Tooltip title="Close">
          <Button
            asIcon
            color="tertiary"
            leftIcon={<CloseIcon className="h-[1.3rem]" strokeWidth={2.8} />}
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
