import { ArrowRightIcon } from "icons";
import { Button, Box, Input, Text, Flex } from "ui";

type ChatInputProps = {
  value: string;
  onChange: (value: string) => void;
  onSend: () => void;
  isLoading: boolean;
};

export const ChatInput = ({
  value,
  onChange,
  onSend,
  isLoading,
}: ChatInputProps) => {
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSend();
    }
  };

  return (
    <Box className="border-t border-gray-100 p-6 dark:border-dark-100">
      <Flex align="end" gap={3}>
        <Box className="flex-1">
          <Input
            disabled={isLoading}
            onChange={(e) => {
              onChange(e.target.value);
            }}
            onKeyDown={handleKeyDown}
            placeholder="Ask Maya anything about your projects..."
            rounded="lg"
            size="lg"
            value={value}
          />
        </Box>
        <Button
          className="h-[3.2rem] w-[3.2rem] p-0"
          disabled={!value.trim() || isLoading}
          onClick={onSend}
          rounded="full"
        >
          <ArrowRightIcon />
        </Button>
      </Flex>
      <Text className="mt-3" color="muted" fontSize="sm">
        Press Enter to send, Shift+Enter for new line
      </Text>
    </Box>
  );
};
