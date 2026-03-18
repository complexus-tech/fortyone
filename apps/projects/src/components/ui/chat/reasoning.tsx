import { useState } from "react";
import { Box, Text } from "ui";
import { cn } from "lib";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";

interface ReasoningProps {
  content: string;
  isStreaming?: boolean;
  className?: string;
}

export const Reasoning = ({
  content,
  isStreaming = false,
  className,
}: ReasoningProps) => {
  const [isExpanded, setIsExpanded] = useState(false);

  if (!content) return null;

  // Don't show anything while still streaming — avoids the brief flash
  // of reasoning tokens when the response completes
  if (isStreaming) return null;

  return (
    <Box className={cn("mb-3", className)}>
      <button
        className="flex items-center gap-1.5"
        onClick={() => setIsExpanded((prev) => !prev)}
        type="button"
      >
        <svg
          className={cn(
            "h-3.5 w-3.5 text-current transition-transform duration-200",
            { "rotate-90": isExpanded },
          )}
          fill="none"
          stroke="currentColor"
          strokeWidth={2.5}
          viewBox="0 0 24 24"
        >
          <path d="M9 5l7 7-7 7" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
        <Text color="muted" className="text-base">
          {isExpanded ? "Hide reasoning" : "Show reasoning"}
        </Text>
      </button>
      {isExpanded && (
        <Box className="prose prose-stone dark:prose-invert mt-2 border-l-2 border-current/10 pl-3 leading-snug opacity-70">
          <Markdown rehypePlugins={[rehypeRaw]} remarkPlugins={[remarkGfm]}>
            {content}
          </Markdown>
        </Box>
      )}
    </Box>
  );
};
