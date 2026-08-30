"use client";

import { LinkIcon } from "icons";
import { Box, Flex, Text } from "ui";
import {
  getRequestAttachments,
  getRequestExternalLinks,
} from "../utils/request-metadata";

export const RequestMetadata = ({
  metadata,
}: {
  metadata: Record<string, unknown>;
}) => {
  const links = getRequestExternalLinks(metadata);
  const attachments = getRequestAttachments(metadata);

  if (links.length === 0 && attachments.length === 0) return null;

  return (
    <>
      {links.length > 0 ? (
        <Box className="border-border mt-5 border-t-[0.5px] pt-4">
          <Text as="h4" className="mb-3" fontWeight="medium">
            External links
          </Text>
          <Box className="space-y-2">
            {links.map((link) => (
              <a
                className="border-border hover:bg-surface-muted flex items-center gap-2 rounded-lg border px-3 py-2 transition"
                href={link.url}
                key={`${link.title}-${link.url}`}
                rel="noopener noreferrer"
                target="_blank"
              >
                <LinkIcon className="h-4 shrink-0" />
                <Text className="line-clamp-1">{link.title}</Text>
              </a>
            ))}
          </Box>
        </Box>
      ) : null}
      {attachments.length > 0 ? (
        <Box className="border-border mt-5 border-t-[0.5px] pt-4">
          <Text as="h4" className="mb-3" fontWeight="medium">
            Attachments
          </Text>
          <Box className="space-y-2">
            {attachments.map((attachment) =>
              attachment.url ? (
                <a
                  className="border-border hover:bg-surface-muted flex items-center gap-2 rounded-lg border px-3 py-2 transition"
                  href={attachment.url}
                  key={`${attachment.name}-${attachment.url}`}
                  rel="noopener noreferrer"
                  target="_blank"
                >
                  <LinkIcon className="h-4 shrink-0" />
                  <Text className="line-clamp-1">{attachment.name}</Text>
                </a>
              ) : (
                <Flex
                  align="center"
                  className="border-border rounded-lg border px-3 py-2"
                  gap={2}
                  key={attachment.name}
                >
                  <LinkIcon className="h-4 shrink-0" />
                  <Text className="line-clamp-1">{attachment.name}</Text>
                </Flex>
              ),
            )}
          </Box>
        </Box>
      ) : null}
    </>
  );
};
