"use client";

import { Box, Flex, Text, Menu } from "ui";
import {
  GitHubIcon,
  LinkIcon,
  MoreHorizontalIcon,
  NewTabIcon,
  UnlinkIcon,
} from "icons";
import { useDeleteStoryGitHubLink } from "@/lib/hooks/github";
import type { StoryGitHubLink } from "@/modules/settings/workspace/integrations/github/types";

const typeLabel = (type: string) => {
  switch (type) {
    case "issue":
      return "Issue synced with GitHub";
    case "pull_request":
      return "Pull request linked";
    case "branch":
      return "Branch linked";
    case "commit":
      return "Commit linked";
    default:
      return "Linked to GitHub";
  }
};

export const GitHubBanner = ({
  links,
  storyId,
}: {
  links: StoryGitHubLink[];
  storyId: string;
}) => {
  const issueLink = links.find((l) => l.externalType === "issue");
  const prLinks = links.filter((l) => l.externalType === "pull_request");
  const primaryLink = issueLink ?? prLinks[0];

  if (!primaryLink) return null;

  return (
    <Box className="mb-3 space-y-2">
      {issueLink && <BannerRow link={issueLink} storyId={storyId} />}
      {prLinks.map((pr) => (
        <BannerRow key={pr.id} link={pr} storyId={storyId} />
      ))}
    </Box>
  );
};

const BannerRow = ({
  link,
  storyId,
}: {
  link: StoryGitHubLink;
  storyId: string;
}) => {
  const { mutate: deleteLink } = useDeleteStoryGitHubLink();
  const label = typeLabel(link.externalType);
  const number = link.githubNumber ? `#${link.githubNumber}` : "";

  return (
    <Flex
      align="center"
      className="border-primary/20 bg-primary/5 rounded-xl border px-4 py-3"
      justify="between"
    >
      <Flex align="center" gap={2}>
        <GitHubIcon className="text-primary h-5 shrink-0" />
        <Text className="text-primary font-medium">
          {label} {number}
        </Text>
      </Flex>
      <Flex align="center" gap={1}>
        <a
          className="text-primary hover:text-primary/80 rounded-md p-1 transition"
          href={link.url}
          rel="noopener noreferrer"
          target="_blank"
          title="Open on GitHub"
        >
          <LinkIcon className="text-current" />
        </a>
        <Menu>
          <Menu.Button>
            <button
              className="text-primary hover:text-primary/80 rounded-md p-1 transition"
              type="button"
            >
              <MoreHorizontalIcon className="h-5 text-current" />
            </button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  window.open(link.url, "_blank", "noopener,noreferrer");
                }}
              >
                <GitHubIcon className="h-5 w-auto" />
                Open on GitHub
              </Menu.Item>
              <Menu.Item
                onSelect={() => {
                  navigator.clipboard.writeText(link.url);
                }}
              >
                <LinkIcon />
                Copy link
              </Menu.Item>
              <Menu.Item
                onSelect={() => {
                  deleteLink({ storyId, linkId: link.id });
                }}
              >
                <UnlinkIcon />
                Unlink
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Flex>
  );
};
