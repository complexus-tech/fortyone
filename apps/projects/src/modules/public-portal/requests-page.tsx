import { ArrowUpDownIcon } from "icons";
import { Box, Flex } from "ui";
import { cn } from "lib";
import { requestFilters, requestStatusMeta } from "./status";
import { PublicPortalShell } from "./portal-shell";
import { PublicRequestCard } from "./request-card";
import { PublicPortalSidebar } from "./sidebar";
import type { PublicPortal, PublicPortalViewer } from "./types";

export const PublicPortalRequestsPage = ({
  portal,
  viewer,
}: {
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => {
  const getStatusCount = (status: (typeof requestFilters)[number]) =>
    portal.requests.filter((request) => request.status === status).length;

  return (
    <PublicPortalShell activeTab="requests" portal={portal} viewer={viewer}>
      <Box className="border-border/60 bg-surface/20 border-b-[0.5px]">
        <Box className="mx-auto flex min-h-16 max-w-[78rem] items-center gap-4 overflow-x-auto px-4 md:px-6">
          <Flex
            align="center"
            className="bg-surface border-border/70 shadow-shadow/30 shrink-0 gap-1 rounded-full border-[0.5px] p-1 shadow-sm"
          >
            <button
              className="bg-state-selected/50 text-foreground border-border shadow-xs dark:bg-state-selected rounded-full border-[0.5px] px-3.5 py-1.5 transition"
              type="button"
            >
              All
              <span className="text-text-muted ml-1.5">
                {portal.requests.length}
              </span>
            </button>
            {requestFilters.map((status) => {
              const meta = requestStatusMeta[status];
              const count = getStatusCount(status);
              return (
                <button
                  className="text-text-muted hover:bg-state-hover hover:text-foreground flex shrink-0 items-center gap-2 rounded-full px-3.5 py-1.5 transition"
                  key={status}
                  type="button"
                >
                  <span
                    className={cn("size-2 rounded-full", meta.dotClassName)}
                  />
                  <span>{meta.label}</span>
                  <span className="text-text-muted/80">{count}</span>
                </button>
              );
            })}
          </Flex>
          <Box className="border-border h-5 shrink-0 border-l-[0.5px]" />
          <Flex
            align="center"
            className="bg-surface border-border/70 shadow-shadow/30 shrink-0 gap-1 rounded-full border-[0.5px] p-1 shadow-sm"
          >
            <button
              className="bg-state-selected/50 text-foreground border-border shadow-xs dark:bg-state-selected flex items-center gap-1.5 rounded-full border-[0.5px] px-3.5 py-1.5 transition"
              type="button"
            >
              <ArrowUpDownIcon className="h-4 text-current" />
              Top
            </button>
            <button
              className="text-text-muted hover:bg-state-hover hover:text-foreground rounded-full px-3.5 py-1.5 transition"
              type="button"
            >
              Newest
            </button>
          </Flex>
        </Box>
      </Box>

      <Box className="mx-auto grid w-full max-w-[78rem] gap-10 px-4 py-8 md:grid-cols-[minmax(0,1fr)_19rem] md:px-6">
        <Box>
          {portal.requests.map((request) => (
            <PublicRequestCard
              key={request.id}
              portal={portal}
              request={request}
            />
          ))}
        </Box>
        <PublicPortalSidebar portal={portal} />
      </Box>
    </PublicPortalShell>
  );
};
