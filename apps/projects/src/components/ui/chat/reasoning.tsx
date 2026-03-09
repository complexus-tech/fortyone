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
        "prose prose-stone dark:prose-invert mb-5 line-clamp-5 animate-pulse pb-1 leading-snug",
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
