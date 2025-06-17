import { Box } from "ui";
import ReactMarkdown from "react-markdown";

type MarkdownProps = {
  content: string;
};

export function Markdown({ content }: MarkdownProps) {
  return (
    <Box className="max-w-none">
      <ReactMarkdown>{content}</ReactMarkdown>
    </Box>
  );
}
