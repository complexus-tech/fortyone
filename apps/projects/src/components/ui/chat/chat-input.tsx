import { ArrowRight2Icon } from "icons";
import { Button, Box, Flex } from "ui";
import { useRef, useEffect } from "react";

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
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
    }
  }, [value]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSend();
    }
  };

  return (
    <Box className="border-t border-gray-100 p-6 dark:border-dark-100">
      <Flex className="items-end gap-0 rounded-[1.25rem] border border-gray-100 bg-gray-50/50 px-4 py-3 dark:border-dark-50 dark:bg-dark-200">
        <textarea
          className="max-h-40 min-h-9 flex-1 resize-none border-none bg-transparent py-2 pr-2 text-xl shadow-none placeholder:text-gray focus:outline-none focus:ring-0 dark:text-white dark:placeholder:text-gray-200/60"
          disabled={isLoading}
          onChange={(e) => {
            onChange(e.target.value);
          }}
          onKeyDown={handleKeyDown}
          placeholder="Ask, suggest, or request for anything..."
          ref={textareaRef}
          rows={1}
          value={value}
        />
        <Button
          asIcon
          className="mb-0.5 ml-1 md:h-11"
          color="invert"
          onClick={onSend}
          rounded="full"
        >
          <ArrowRight2Icon className="text-white dark:text-dark" />
        </Button>
      </Flex>
    </Box>
  );
};
