"use client";

import { CopyIcon, ShareIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { toast } from "sonner";
import type { PublicPortal } from "./types";
import { NewFeedbackButton } from "./feedback-controls";

const copyLink = async () => {
  try {
    await navigator.clipboard.writeText(window.location.href);
    toast.success("Link copied to clipboard");
  } catch {
    toast.error("Unable to copy link");
  }
};

const shareLink = async (portal: PublicPortal) => {
  const share = Reflect.get(navigator, "share") as
    | Navigator["share"]
    | undefined;
  if (typeof share !== "function") {
    toast.error("Sharing is not supported in this browser");
    return;
  }

  try {
    await share.call(navigator, {
      title: portal.name,
      text: portal.description || `${portal.workspace.name} feedback`,
      url: window.location.href,
    });
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") return;
    toast.error("Unable to share link");
  }
};

export const PublicPortalSidebar = ({ portal }: { portal: PublicPortal }) => {
  return (
    <aside className="space-y-8">
      <NewFeedbackButton portal={portal} />

      <Box className="border-border bg-surface shadow-shadow/40 rounded-xl border-[0.5px] p-2 shadow-sm">
        <Text
          className="px-3 py-2 text-[0.8rem] tracking-[0.12em] uppercase"
          color="muted"
        >
          Boards
        </Text>
        <Box className="bg-state-selected rounded-lg px-3 py-2.5">
          <Flex align="center" gap={2}>
            <span className="bg-text-muted size-2 rounded-full" />
            <Text fontWeight="semibold">All Feedback</Text>
          </Flex>
        </Box>
        {portal.boards.map((board) => (
          <Flex
            align="center"
            className="hover:bg-state-hover rounded-lg px-3 py-2.5 transition"
            gap={2}
            key={board.id}
          >
            <span className={`${board.colorClassName} size-2 rounded-full`} />
            <Text color="muted">{board.name}</Text>
          </Flex>
        ))}
      </Box>

      <Box className="border-border bg-surface shadow-shadow/40 rounded-xl border-[0.5px] p-2 shadow-sm">
        <Text className="px-2 py-2" fontWeight="semibold">
          Actions
        </Text>
        <Flex direction="column" gap={1}>
          <Button
            className="w-full justify-start px-2"
            color="tertiary"
            leftIcon={<CopyIcon className="h-4" />}
            onClick={() => {
              void copyLink();
            }}
            variant="naked"
          >
            Copy link
          </Button>
          <Button
            className="w-full justify-start px-2"
            color="tertiary"
            leftIcon={<ShareIcon className="h-4" />}
            onClick={() => {
              void shareLink(portal);
            }}
            variant="naked"
          >
            Share
          </Button>
        </Flex>
      </Box>
    </aside>
  );
};
