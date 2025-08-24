import { Box } from "ui";
import { cn } from "lib";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";

interface ReasoningProps {
  content: string;
  isStreaming?: boolean;
  className?: string;
  defaultExpanded?: boolean;
  title?: string;
}

export const Reasoning = ({ content, isStreaming = false }: ReasoningProps) => {
  if (!content || !isStreaming) {
    return null;
  }

  return (
    <Box
      className={cn(
        "prose prose-stone mb-6 line-clamp-2 animate-pulse leading-none dark:prose-invert",
        {
          "opacity-80": isStreaming,
        },
      )}
    >
      <Markdown rehypePlugins={[rehypeRaw]} remarkPlugins={[remarkGfm]}>
        {content}
      </Markdown>
    </Box>
  );
};
