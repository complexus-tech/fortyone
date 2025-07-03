import { Button, Box, Flex } from "ui";
import type { ChangeEvent } from "react";
import { useRef, useEffect } from "react";
import { PlusIcon } from "icons";

type ChatInputProps = {
  value: string;
  onChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  onSend: () => void;
  onStop: () => void;
  isLoading: boolean;
};

const SendIcon = () => {
  return (
    <svg
      className="h-6 w-auto text-white"
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

const StopIcon = () => {
  return (
    <svg
      className="h-5 w-auto text-white"
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M12.0436 3.25C13.6463 3.24999 14.9086 3.24998 15.913 3.35586C16.9399 3.4641 17.7833 3.68971 18.5113 4.19945C19.0129 4.55072 19.4493 4.98706 19.8005 5.48872C20.3103 6.21671 20.5359 7.06008 20.6441 8.08697C20.75 9.0914 20.75 10.3537 20.75 11.9564V12.0436C20.75 13.6463 20.75 14.9086 20.6441 15.913C20.5359 16.9399 20.3103 17.7833 19.8005 18.5113C19.4493 19.0129 19.0129 19.4493 18.5113 19.8005C17.7833 20.3103 16.9399 20.5359 15.913 20.6441C14.9086 20.75 13.6463 20.75 12.0436 20.75H11.9564C10.3537 20.75 9.0914 20.75 8.08697 20.6441C7.06008 20.5359 6.21671 20.3103 5.48872 19.8005C4.98706 19.4493 4.55072 19.0129 4.19945 18.5113C3.68971 17.7833 3.4641 16.9399 3.35586 15.913C3.24998 14.9086 3.24999 13.6463 3.25 12.0436V11.9564C3.24999 10.3537 3.24998 9.0914 3.35586 8.08697C3.4641 7.06008 3.68971 6.21671 4.19945 5.48872C4.55072 4.98706 4.98706 4.55072 5.48872 4.19945C6.21671 3.68971 7.06008 3.4641 8.08697 3.35586C9.0914 3.24998 10.3537 3.24999 11.9564 3.25H12.0436Z"
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
  onStop,
}: ChatInputProps) => {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

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
    <Box className="px-6 pb-6">
      <Box className="rounded-[1.25rem] border border-gray-100 bg-gray-50/80 py-2 dark:border-dark-50 dark:bg-dark-100/70">
        <textarea
          autoFocus
          className="max-h-40 min-h-9 w-full flex-1 resize-none border-none bg-transparent px-5 py-2 text-lg shadow-none placeholder:text-gray focus:outline-none focus:ring-0 dark:text-white dark:placeholder:text-gray-200/60"
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
            color="tertiary"
            onClick={onSend}
            rounded="full"
            variant="naked"
          >
            <PlusIcon />
          </Button>
          <Button
            asIcon
            className="mb-0.5 md:h-11"
            onClick={() => {
              if (isLoading) {
                onStop();
              } else {
                onSend();
              }
            }}
            rounded="full"
          >
            {isLoading ? <StopIcon /> : <SendIcon />}
          </Button>
        </Flex>
      </Box>
    </Box>
  );
};
