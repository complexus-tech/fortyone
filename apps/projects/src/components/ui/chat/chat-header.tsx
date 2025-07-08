import { CloseIcon, MaximizeIcon, MinimizeIcon, PlusIcon } from "icons";
import { Flex, Button } from "ui";
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
      <Button
        asIcon
        color="tertiary"
        disabled
        leftIcon={
          <PlusIcon
            className="h-[1.4rem] text-dark/70 dark:text-gray-200"
            strokeWidth={2.8}
          />
        }
        size="sm"
        variant="naked"
      >
        <span className="sr-only">New chat</span>
      </Button>
      <Flex align="center" gap={4}>
        {isDesktop ? (
          <Button
            asIcon
            color="tertiary"
            leftIcon={
              isFullScreen ? (
                <MinimizeIcon
                  className="h-[1.4rem] text-dark/70 dark:text-gray-200"
                  strokeWidth={2.8}
                />
              ) : (
                <MaximizeIcon
                  className="h-[1.4rem] text-dark/70 dark:text-gray-200"
                  strokeWidth={2.8}
                />
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
          leftIcon={
            <CloseIcon
              className="h-[1.4rem] text-dark/70 dark:text-gray-200"
              strokeWidth={2.8}
            />
          }
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
