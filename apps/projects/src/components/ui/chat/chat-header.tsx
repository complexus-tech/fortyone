import { CloseIcon, MaximizeIcon, PlusIcon } from "icons";
import { Flex, Button } from "ui";

export const ChatHeader = ({
  setIsOpen,
}: {
  setIsOpen: (isOpen: boolean) => void;
}) => {
  return (
    <Flex align="center" justify="between">
      <Button
        asIcon
        color="tertiary"
        leftIcon={
          <PlusIcon
            className="h-[1.4rem] text-dark/70 dark:text-gray-200"
            strokeWidth={2.8}
          />
        }
        size="sm"
        variant="naked"
      >
        <span className="sr-only">New</span>
      </Button>
      <Flex align="center" gap={4}>
        <Button
          asIcon
          color="tertiary"
          leftIcon={
            <MaximizeIcon
              className="h-[1.4rem] text-dark/70 dark:text-gray-200"
              strokeWidth={2.8}
            />
          }
          size="sm"
          variant="naked"
        >
          <span className="sr-only">Maximize</span>
        </Button>
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
