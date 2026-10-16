import { useQuery } from "@tanstack/react-query";
import { Box, Button, Flex, Text } from "ui";
import { LoadingIcon } from "icons";
import type { StoryAttachment } from "../types";
import { readAttachmentText } from "./attachment-text";

export const AttachmentTextPreview = ({ file }: { file: StoryAttachment }) => {
  const preview = useQuery({
    queryKey: ["comment-attachment-text", file.id, file.url],
    queryFn: ({ signal }) => readAttachmentText(file, signal),
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
  if (preview.isPending)
    return (
      <Flex align="center" className="h-[50dvh]" gap={2} justify="center">
        <LoadingIcon className="size-4 animate-spin" />
        <Text color="muted">Loading preview…</Text>
      </Flex>
    );
  if (preview.isError)
    return (
      <Flex
        align="center"
        className="h-[50dvh]"
        direction="column"
        gap={3}
        justify="center"
      >
        <Text color="muted">{preview.error.message}</Text>
        <Button
          color="tertiary"
          onClick={() => {
            void preview.refetch();
          }}
          size="sm"
          variant="outline"
        >
          Retry preview
        </Button>
      </Flex>
    );
  return (
    <Box>
      {preview.data.truncated ? (
        <Text
          className="border-border border-b px-4 py-2 text-sm"
          color="muted"
        >
          Showing the first 256 KB. Download to view the complete file.
        </Text>
      ) : null}
      <pre
        aria-label={`${file.filename} text preview`}
        className="h-[50dvh] overflow-auto p-4 font-mono text-sm leading-6"
      >
        {preview.data.text || "This file is empty."}
      </pre>
    </Box>
  );
};
