import {
  AiIcon,
  CloseIcon,
  MaximizeIcon,
  MinimizeIcon,
  PlusIcon,
  PreferencesIcon,
} from "icons";
import { Flex, Button, Text } from "ui";
import { useMediaQuery } from "@/hooks";

export const ChatHeader = ({
  setIsOpen,
  isFullScreen,
  setIsFullScreen,
}: {
  setIsOpen: (isOpen: boolean) => void;
  isFullScreen: boolean;
  setIsFullScreen: (isFullScreen: boolean) => void;
}) => {
  const isDesktop = useMediaQuery("(min-width: 768px)");
  return (
    <Flex align="center" justify="between">
      <Flex align="center" gap={2}>
        <AiIcon className="h-6" />
        <Text>Maya is your AI assistant</Text>
      </Flex>
      <Flex align="center" gap={2}>
        <Button
          asIcon
          color="tertiary"
          disabled
          leftIcon={<PlusIcon strokeWidth={2.8} />}
          size="sm"
          variant="naked"
        >
          <span className="sr-only">New chat</span>
        </Button>
        <Button
          asIcon
          color="tertiary"
          disabled
          leftIcon={<PreferencesIcon />}
          size="sm"
          variant="naked"
        >
          <span className="sr-only">New chat</span>
        </Button>
        {isDesktop ? (
          <Button
            asIcon
            color="tertiary"
            leftIcon={
              isFullScreen ? (
                <MinimizeIcon strokeWidth={2.8} />
              ) : (
                <MaximizeIcon strokeWidth={2.8} />
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
        ) : null}
        <Button
          asIcon
          color="tertiary"
          leftIcon={<CloseIcon strokeWidth={2.8} />}
          onClick={() => {
            setIsOpen(false);
          }}
          size="sm"
          variant="naked"
        >
          <span className="sr-only">Close</span>
        </Button>
      </Flex>
    </Flex>
  );
};
