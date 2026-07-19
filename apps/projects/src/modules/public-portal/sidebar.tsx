"use client";

import { CheckIcon, CopyIcon, ShareIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import type { PublicPortal, PublicPortalViewer } from "./types";
import { NewFeedbackButton } from "./feedback-controls";
import { getPortalLoginUrl } from "./utils";

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

export const PublicPortalActions = ({ portal }: { portal: PublicPortal }) => (
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
);

export const PublicPortalSidebar = ({
  onBoardSelect,
  portal,
  selectedBoardId,
  viewer,
}: {
  onBoardSelect: (boardId?: string) => void;
  portal: PublicPortal;
  selectedBoardId?: string;
  viewer?: PublicPortalViewer | null;
}) => {
  return (
    <aside className="space-y-8 md:min-h-0 md:overflow-y-auto">
      {viewer ? (
        <NewFeedbackButton portal={portal} />
      ) : (
        <Button
          className="h-12 w-full justify-center text-[1rem]"
          color="invert"
          href={getPortalLoginUrl(portal, "feedback")}
          size="lg"
        >
          Login to submit feedback
        </Button>
      )}

      <Box className="border-border bg-surface shadow-shadow/40 rounded-xl border-[0.5px] p-2 shadow-sm">
        <Text
          className="px-3 py-2 text-[0.8rem] tracking-[0.12em] uppercase"
          color="muted"
        >
          Boards
        </Text>
        <button
          aria-pressed={!selectedBoardId}
          className={cn(
            "hover:bg-state-hover flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-left transition",
            { "bg-state-hover": !selectedBoardId },
          )}
          onClick={() => {
            onBoardSelect(undefined);
          }}
          type="button"
        >
          <span className="bg-text-muted size-2 rounded-full" />
          <Text fontWeight={!selectedBoardId ? "semibold" : "normal"}>
            All boards
          </Text>
          {!selectedBoardId ? (
            <CheckIcon className="ml-auto h-4 w-auto" />
          ) : null}
        </button>
        {portal.boards.map((board) => (
          <button
            aria-pressed={selectedBoardId === board.id}
            className={cn(
              "hover:bg-state-hover flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-left transition",
              { "bg-state-hover": selectedBoardId === board.id },
            )}
            key={board.id}
            onClick={() => {
              onBoardSelect(board.id);
            }}
            type="button"
          >
            <span className={`${board.colorClassName} size-2 rounded-full`} />
            <Text
              color={selectedBoardId === board.id ? undefined : "muted"}
              fontWeight={selectedBoardId === board.id ? "semibold" : "normal"}
            >
              {board.name}
            </Text>
            {selectedBoardId === board.id ? (
              <CheckIcon className="ml-auto h-4 w-auto" />
            ) : null}
          </button>
        ))}
      </Box>

      <PublicPortalActions portal={portal} />
    </aside>
  );
};
