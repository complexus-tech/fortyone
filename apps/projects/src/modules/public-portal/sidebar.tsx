import { CopyIcon, PlusIcon, ShareIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import type { PublicPortal } from "./types";

export const PublicPortalSidebar = ({ portal }: { portal: PublicPortal }) => (
  <aside className="space-y-8">
    <Button
      className="h-12 w-full justify-center text-[1rem]"
      color="invert"
      leftIcon={<PlusIcon className="h-4 text-current" />}
      rounded="full"
      size="lg"
    >
      New Request
    </Button>

    <Box className="border-border bg-surface shadow-shadow/40 rounded-3xl border-[0.5px] p-2 shadow-sm">
      <Text
        className="px-3 py-2 text-[0.8rem] uppercase tracking-[0.12em]"
        color="muted"
      >
        Boards
      </Text>
      <Box className="bg-state-selected/50 dark:bg-state-selected rounded-full px-3 py-2.5">
        <Flex align="center" gap={2}>
          <span className="bg-text-muted size-2 rounded-full" />
          <Text fontWeight="semibold">All Requests</Text>
        </Flex>
      </Box>
      {portal.boards.map((board) => (
        <Flex
          align="center"
          className="hover:bg-state-hover rounded-full px-3 py-2.5 transition"
          gap={2}
          key={board.id}
        >
          <span className={`${board.colorClassName} size-2 rounded-full`} />
          <Text color="muted">{board.name}</Text>
        </Flex>
      ))}
    </Box>

    <Box className="border-border bg-surface shadow-shadow/40 rounded-3xl border-[0.5px] p-4 shadow-sm">
      <Text className="mb-4" fontWeight="semibold">
        Actions
      </Text>
      <Flex className="text-text-muted gap-4" direction="column">
        <Flex align="center" gap={2}>
          <CopyIcon className="h-4" />
          <span>Copy link</span>
        </Flex>
        <Flex align="center" gap={2}>
          <ShareIcon className="h-4" />
          <span>Share</span>
        </Flex>
      </Flex>
    </Box>
  </aside>
);
