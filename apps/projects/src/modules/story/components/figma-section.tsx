"use client";

import { Badge, Box, Button, Flex, Menu, Text, TimeAgo } from "ui";
import { LinkIcon, MoreHorizontalIcon, RefreshIcon, UnlinkIcon } from "icons";
import { toast } from "sonner";
import {
  useDeleteFigmaStoryLink,
  useRefreshFigmaStoryLink,
  useStoryFigmaLinks,
} from "@/lib/hooks/figma";
import type { StoryFigmaLink } from "@/modules/settings/workspace/integrations/figma/types";
import { FigmaIcon } from "@/modules/settings/workspace/integrations/figma/icon";

const statusLabel = (status: StoryFigmaLink["devStatus"]) => {
  switch (status) {
    case "READY_FOR_DEV":
      return "Ready for development";
    case "COMPLETED":
      return "Completed in Figma";
    default:
      return null;
  }
};

const DesignSyncLabel = ({ link }: { link: StoryFigmaLink }) => {
  if (link.unavailableAt) return <>Preview unavailable</>;
  if (link.artifact.lastModified) {
    return <TimeAgo timestamp={link.artifact.lastModified} />;
  }
  return <>Linked design</>;
};

const FigmaDesignCard = ({
  link,
  storyId,
}: {
  link: StoryFigmaLink;
  storyId: string;
}) => {
  const removeLink = useDeleteFigmaStoryLink();
  const refreshLink = useRefreshFigmaStoryLink();
  const status = statusLabel(link.devStatus);

  return (
    <Box className="border-border bg-surface overflow-hidden rounded-xl border">
      {link.artifact.thumbnailUrl ? (
        <a
          aria-label={`Open ${link.artifact.nodeName ?? link.artifact.fileName} in Figma`}
          className="bg-surface-muted block aspect-[2.4/1] bg-cover bg-center transition-opacity hover:opacity-90"
          href={link.artifact.canonicalUrl}
          rel="noopener noreferrer"
          style={{
            backgroundImage: `url("${link.artifact.thumbnailUrl.replaceAll('"', "%22")}")`,
          }}
          target="_blank"
        />
      ) : null}
      <Flex align="center" className="px-4 py-3" justify="between">
        <Flex align="center" className="min-w-0" gap={3}>
          <Flex
            align="center"
            className="bg-surface-muted size-9 shrink-0 rounded-lg"
            justify="center"
          >
            <FigmaIcon className="h-5 w-auto" />
          </Flex>
          <Box className="min-w-0">
            <Flex align="center" gap={2}>
              <Text className="truncate font-medium">
                {link.artifact.nodeName ?? link.artifact.fileName}
              </Text>
              {status ? <Badge color="secondary">{status}</Badge> : null}
            </Flex>
            <Text className="truncate" color="muted">
              {link.artifact.nodeName ? `${link.artifact.fileName} · ` : ""}
              <DesignSyncLabel link={link} />
            </Text>
          </Box>
        </Flex>
        <Flex align="center" gap={1}>
          <Button
            color="tertiary"
            onClick={() =>
              window.open(
                link.artifact.canonicalUrl,
                "_blank",
                "noopener,noreferrer",
              )
            }
            size="sm"
            variant="naked"
          >
            Open in Figma
          </Button>
          <Menu>
            <Menu.Button>
              <button
                aria-label="More Figma design actions"
                className="text-icon hover:bg-surface-muted rounded-md p-1.5"
                type="button"
              >
                <MoreHorizontalIcon className="h-5" />
              </button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item
                  disabled={refreshLink.isPending}
                  onSelect={() => {
                    refreshLink.mutate({ storyId, linkId: link.id });
                  }}
                >
                  <RefreshIcon />
                  Refresh preview
                </Menu.Item>
                <Menu.Item
                  onSelect={() => {
                    void navigator.clipboard.writeText(
                      link.artifact.canonicalUrl,
                    );
                    toast.success("Figma link copied");
                  }}
                >
                  <LinkIcon />
                  Copy link
                </Menu.Item>
                <Menu.Item
                  onSelect={() => {
                    removeLink.mutate({ storyId, linkId: link.id });
                  }}
                >
                  <UnlinkIcon />
                  Remove design
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      </Flex>
    </Box>
  );
};

export const FigmaSection = ({ storyId }: { storyId: string }) => {
  const { data: links = [] } = useStoryFigmaLinks(storyId);
  if (links.length === 0) return null;

  return (
    <Box className="mt-4 space-y-2">
      <Flex align="center" gap={2}>
        <FigmaIcon className="h-4 w-auto" />
        <Text className="font-semibold">Design</Text>
      </Flex>
      {links.map((link) => (
        <FigmaDesignCard key={link.id} link={link} storyId={storyId} />
      ))}
    </Box>
  );
};
