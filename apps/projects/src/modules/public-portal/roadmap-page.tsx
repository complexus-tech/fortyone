import { ArrowUpIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { roadmapStatuses, requestStatusMeta } from "./status";
import { PublicPortalShell } from "./portal-shell";
import type { PublicPortal, PublicPortalViewer } from "./types";
import { getBoard } from "./utils";

const roadmapLabels = {
  planned: {
    title: "Planned",
    description: "Committed and queued",
  },
  in_progress: {
    title: "In Progress",
    description: "Actively being delivered",
  },
  completed: {
    title: "Done",
    description: "Recently completed",
  },
};

export const PublicPortalRoadmapPage = ({
  portal,
  viewer,
}: {
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => (
  <PublicPortalShell activeTab="roadmap" portal={portal} viewer={viewer}>
    <Box className="mx-auto grid min-h-[calc(100dvh-4rem)] w-full max-w-[78rem] gap-5 px-4 py-8 md:grid-cols-3 md:px-6">
      {roadmapStatuses.map((status) => {
        const items = portal.requests.filter(
          (request) => request.status === status,
        );
        const column = roadmapLabels[status];

        return (
          <Box className="min-h-[32rem]" key={status}>
            <Flex align="center" className="items-baseline" gap={2}>
              <Text className="text-[1.25rem]" fontWeight="semibold">
                {column.title}
              </Text>
              <Text color="muted">{items.length}</Text>
            </Flex>
            <Text className="mt-1" color="muted">
              {column.description}
            </Text>
            <Box className="border-border/70 mt-5 min-h-[26rem] rounded-3xl border-[0.5px] border-dashed p-3">
              {items.length === 0 ? (
                <Flex align="center" className="h-64" justify="center">
                  <Text color="muted">Nothing here yet</Text>
                </Flex>
              ) : (
                <Flex direction="column" gap={3}>
                  {items.map((request) => {
                    const board = getBoard(portal, request.boardId);
                    const meta = requestStatusMeta[request.status];
                    return (
                      <Box
                        className="border-border bg-surface rounded-3xl border-[0.5px] p-4"
                        key={request.id}
                      >
                        {board ? (
                          <Flex align="center" className="mb-4" gap={2}>
                            <span
                              className={`${meta.dotClassName} size-2.5 rounded-full`}
                            />
                            <Text color="muted">{board.name}</Text>
                          </Flex>
                        ) : null}
                        <Text
                          className="line-clamp-2 text-[1.05rem]"
                          fontWeight="semibold"
                        >
                          {request.title}
                        </Text>
                        {request.roadmapSummary ? (
                          <Text className="mt-2 line-clamp-2" color="muted">
                            {request.roadmapSummary}
                          </Text>
                        ) : null}
                        <Flex align="center" className="mt-6" justify="between">
                          <Text color="muted">{request.authorName}</Text>
                          <Button
                            className="h-7 rounded-xl border-none px-2.5"
                            color="tertiary"
                            leftIcon={<ArrowUpIcon className="h-3.5" />}
                            size="xs"
                          >
                            {request.voteCount}
                          </Button>
                        </Flex>
                      </Box>
                    );
                  })}
                </Flex>
              )}
            </Box>
          </Box>
        );
      })}
    </Box>
  </PublicPortalShell>
);
