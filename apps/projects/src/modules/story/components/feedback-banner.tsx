"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { CopyIcon, LinkIcon, MoreHorizontalIcon, RequestsIcon } from "icons";
import { Box, Flex, Menu, Text, TimeAgo } from "ui";
import { useWorkspacePath } from "@/hooks";
import type { StoryFeedbackLink } from "@/modules/team-feedback/types";

const relationshipLabel = (relationship: StoryFeedbackLink["relationship"]) =>
  relationship === "created_from"
    ? "Created from feedback"
    : "Linked to feedback";

export const FeedbackBanner = ({ links }: { links: StoryFeedbackLink[] }) => (
  <Box className="mb-3 space-y-2">
    {links.map((link) => (
      <FeedbackBannerRow key={link.id} link={link} />
    ))}
  </Box>
);

const FeedbackBannerRow = ({ link }: { link: StoryFeedbackLink }) => {
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const feedbackHref = withWorkspace(
    `/teams/${link.teamId}/feedback/${link.itemId}`,
  );

  const copyFeedbackLink = () =>
    navigator.clipboard.writeText(
      new URL(feedbackHref, window.location.origin).toString(),
    );

  return (
    <Flex
      align="center"
      className="border-primary/20 bg-primary/5 rounded-xl border px-4 py-2"
      justify="between"
    >
      <Link className="min-w-0" href={feedbackHref} title="Open feedback">
        <Flex align="center" className="min-w-0" gap={2}>
          <RequestsIcon className="text-primary h-5 shrink-0" />
          <Box className="min-w-0">
            <Text color="primary" fontWeight="medium">
              {relationshipLabel(link.relationship)}
            </Text>
            <Text className="line-clamp-1" color="muted">
              {link.feedbackTitle}
              <span aria-hidden="true"> · </span>
              Linked <TimeAgo timestamp={link.createdAt} />
            </Text>
          </Box>
        </Flex>
      </Link>
      <Flex align="center" className="shrink-0" gap={1}>
        <Link
          aria-label="Open feedback"
          className="text-primary hover:text-primary/80 rounded-md p-1 transition"
          href={feedbackHref}
          title="Open feedback"
        >
          <LinkIcon className="text-current" />
        </Link>
        <Menu>
          <Menu.Button>
            <button
              aria-label="More feedback link actions"
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
                  router.push(feedbackHref);
                }}
              >
                <RequestsIcon className="h-5 w-auto" />
                Open feedback
              </Menu.Item>
              <Menu.Item onSelect={copyFeedbackLink}>
                <CopyIcon />
                Copy link
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Flex>
  );
};
