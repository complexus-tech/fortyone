import { Button, Box, Flex } from "ui";
import type { ChangeEvent } from "react";
import { useRef, useEffect } from "react";
import { PlusIcon } from "icons";

type ChatInputProps = {
  value: string;
  onChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  onSend: () => void;
  isLoading: boolean;
};

const SendIcon = () => {
  return (
    <svg
      className="h-6 w-auto"
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
      // focus on the textarea
      textareaRef.current?.focus();
    }
  };

  return (
    <Box className="p-6">
      <Box className="rounded-[1.25rem] border border-gray-100 bg-gray-50/50 py-2 dark:border-dark-50 dark:bg-dark-300">
        <textarea
          className="max-h-40 min-h-9 w-full flex-1 resize-none border-none bg-transparent px-5 py-2 text-lg shadow-none placeholder:text-gray focus:outline-none focus:ring-0 dark:text-white dark:placeholder:text-gray-200/60"
          disabled={isLoading}
          onChange={onChange}
          onKeyDown={handleKeyDown}
          placeholder="Ask, suggest, or request for anything..."
          ref={textareaRef}
          rows={1}
          value={value}
        />
        <Flex align="center" className="pl-2 pr-3" gap={2} justify="between">
          <Button
            asIcon
            className="mb-0.5 dark:hover:bg-dark-50 md:h-11"
            // color="invert"
            onClick={onSend}
            rounded="full"
            variant="naked"
          >
            <PlusIcon />
          </Button>
          <Button
            asIcon
            className="mb-0.5 md:h-11"
            // color="invert"
            onClick={onSend}
            rounded="full"
          >
            <SendIcon />
          </Button>
        </Flex>
      </Box>
    </Box>
  );
};
