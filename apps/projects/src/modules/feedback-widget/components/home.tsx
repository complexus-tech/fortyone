import { EditIcon, RequestsIcon, RoadmapIcon } from "icons";
import { Box, Flex, Text } from "ui";
import { cn } from "lib";
import type {
  PublicPortalUpdate,
  PublicRequest,
  PublicRequestStatus,
} from "@/shared/feedback-widget/types";
import { feedbackRequestStatusMeta as requestStatusMeta } from "@/shared/feedback-widget/status";
import type { WidgetRoadmap } from "./types";
import { roadmapSections, statusAccent } from "./utils";
import { FeedbackRow, VoteButton } from "./widget-ui";

const HomeSectionHeader = ({
  actionLabel,
  icon: Icon,
  onAction,
  title,
}: {
  actionLabel: string;
  icon: typeof RequestsIcon;
  onAction: () => void;
  title: string;
}) => (
  <Flex align="center" className="px-5 pt-6 pb-2" justify="between">
    <Flex align="center" className="min-w-0 gap-2">
      <Icon className="text-text-muted h-4 shrink-0" />
      <Text
        className="truncate text-[11px] tracking-[0.09em] uppercase"
        fontWeight="semibold"
      >
        {title}
      </Text>
    </Flex>
    <button
      className="text-text-muted hover:text-foreground focus-visible:ring-ring shrink-0 rounded-lg px-2 py-1 text-[10px] font-normal tracking-[0.09em] uppercase transition-colors focus-visible:ring-2 focus-visible:outline-none"
      onClick={onAction}
      type="button"
    >
      {actionLabel}
    </button>
  </Flex>
);

export const RoadmapGroupHeader = ({
  count,
  label,
  status,
}: {
  count: number;
  label: string;
  status: PublicRequestStatus;
}) => (
  <Flex align="center" className="relative py-3 pr-1 pl-6" justify="between">
    <span
      className={cn(
        "ring-background absolute top-1/2 -left-1.5 size-3 -translate-y-1/2 rounded-sm ring-4",
        statusAccent(status),
      )}
    />
    <Text className="text-[15px]" fontWeight="semibold">
      {label}
    </Text>
    <Text className="text-[11px] tabular-nums" color="muted">
      {String(count).padStart(2, "0")}
    </Text>
  </Flex>
);

const HomeRoadmapRow = ({
  isVoting,
  isWriteLocked,
  onOpen,
  onVote,
  request,
}: {
  isVoting: boolean;
  isWriteLocked: boolean;
  onOpen: () => void;
  onVote: () => void;
  request: PublicRequest;
}) => {
  const status = requestStatusMeta[request.status];
  return (
    <div className="hover:bg-state-hover/35 relative grid grid-cols-[minmax(0,1fr)_auto] items-start gap-3 py-3 pr-1 pl-6 transition-colors">
      <span
        className={cn(
          "ring-background absolute top-5 -left-1 size-2.5 rounded-sm ring-4",
          statusAccent(request.status),
        )}
      />
      <button
        className="focus-visible:ring-ring min-w-0 text-left focus-visible:ring-2 focus-visible:outline-none"
        onClick={onOpen}
        type="button"
      >
        <Box className="min-w-0">
          <Text
            className="line-clamp-1 text-[13px] leading-5"
            fontWeight="semibold"
          >
            {request.title}
          </Text>
          {request.description ? (
            <Text
              className="mt-1 line-clamp-2 text-[12px] leading-5"
              color="muted"
            >
              {request.description}
            </Text>
          ) : null}
          <Flex align="center" className="mt-2 gap-2 text-[12px]">
            <Text className="font-medium" color="muted">
              {status.label}
            </Text>
            <span className="text-text-muted">&bull;</span>
            <Text className="truncate font-medium" color="muted">
              {request.authorName || "Anonymous"}
            </Text>
          </Flex>
        </Box>
      </button>
      <VoteButton
        disabled={isWriteLocked}
        isPending={isVoting}
        onClick={onVote}
        request={request}
      />
    </div>
  );
};

export const WidgetHome = ({
  feedback,
  isWriteLocked,
  latestUpdate,
  onOpenFeedback,
  onOpenRoadmap,
  onOpenUpdate,
  onOpenRequest,
  onShareFeedback,
  onVote,
  roadmap,
  votingRequestId,
}: {
  feedback: PublicRequest[];
  isWriteLocked: boolean;
  latestUpdate?: PublicPortalUpdate;
  onOpenFeedback: () => void;
  onOpenRoadmap: () => void;
  onOpenUpdate: (update: PublicPortalUpdate) => void;
  onOpenRequest: (request: PublicRequest) => void;
  onShareFeedback: () => void;
  onVote: (request: PublicRequest) => void;
  roadmap: WidgetRoadmap;
  votingRequestId: string | null;
}) => {
  const popularFeedback = feedback.slice(0, 3);
  const homeRoadmapGroups = roadmapSections
    .map((section) => ({
      ...section,
      items: roadmap[section.status].slice(0, 2),
      total: roadmap[section.status].length,
    }))
    .filter((section) => section.items.length > 0);

  return (
    <Box className="pb-4">
      <section aria-labelledby="widget-home-feedback">
        <div id="widget-home-feedback">
          <HomeSectionHeader
            actionLabel="See all"
            icon={RequestsIcon}
            onAction={onOpenFeedback}
            title="Popular feedback"
          />
        </div>
        {popularFeedback.length > 0 ? (
          popularFeedback.map((request) => (
            <FeedbackRow
              isVoting={votingRequestId === request.id}
              isWriteLocked={isWriteLocked}
              key={request.id}
              onOpen={() => {
                onOpenRequest(request);
              }}
              onVote={() => {
                onVote(request);
              }}
              request={request}
            />
          ))
        ) : (
          <Text className="block px-5 py-5 text-[12px]" color="muted">
            Feedback shared by customers will appear here.
          </Text>
        )}
      </section>

      <Box className="border-border/60 border-y p-5">
        <button
          className="border-border bg-state-hover/40 hover:bg-state-hover/55 focus-visible:ring-ring flex w-full items-center gap-4 rounded-lg border p-4 text-left transition-colors focus-visible:ring-2 focus-visible:outline-none"
          onClick={onShareFeedback}
          type="button"
        >
          <Box className="min-w-0 flex-1">
            <Text className="text-[15px] leading-5" fontWeight="semibold">
              Help shape what comes next
            </Text>
            <Text className="mt-1 text-[12px] leading-5" color="muted">
              Share an idea, report a problem, or suggest an improvement.
            </Text>
          </Box>
          <span className="bg-foreground text-background inline-flex h-8 shrink-0 items-center gap-1.5 rounded-xl px-3 text-[11px] font-semibold shadow-sm">
            <EditIcon className="h-3.5 text-current" />
            Share
          </span>
        </button>
      </Box>

      {latestUpdate ? (
        <button
          className="border-border/60 hover:bg-state-hover/35 focus-visible:ring-ring w-full border-b px-5 py-6 text-left transition-colors focus-visible:ring-2 focus-visible:outline-none"
          onClick={() => {
            onOpenUpdate(latestUpdate);
          }}
          type="button"
        >
          <Text
            className="text-[10px] tracking-[0.1em] uppercase"
            color="muted"
            fontWeight="semibold"
          >
            Latest update · {latestUpdate.publishedAtLabel}
          </Text>
          <Text className="mt-2 text-[18px] leading-6" fontWeight="semibold">
            {latestUpdate.title}
          </Text>
          <Text
            className="mt-2 line-clamp-3 text-[12px] leading-5"
            color="muted"
          >
            {latestUpdate.summary || latestUpdate.body}
          </Text>
        </button>
      ) : null}

      <section aria-labelledby="widget-home-roadmap">
        <div id="widget-home-roadmap">
          <HomeSectionHeader
            actionLabel="See roadmap"
            icon={RoadmapIcon}
            onAction={onOpenRoadmap}
            title="On the roadmap"
          />
        </div>
        {homeRoadmapGroups.length > 0 ? (
          <Box className="border-border/70 mr-5 ml-6 border-l border-dashed">
            {homeRoadmapGroups.map((section) => (
              <Box key={section.status}>
                <RoadmapGroupHeader
                  count={section.total}
                  label={section.label}
                  status={section.status}
                />
                {section.items.map((request) => (
                  <HomeRoadmapRow
                    isVoting={votingRequestId === request.id}
                    isWriteLocked={isWriteLocked}
                    key={request.id}
                    onOpen={() => {
                      onOpenRequest(request);
                    }}
                    onVote={() => {
                      onVote(request);
                    }}
                    request={request}
                  />
                ))}
              </Box>
            ))}
          </Box>
        ) : (
          <Text className="block px-5 py-5 text-[12px]" color="muted">
            Planned and in-progress ideas will appear here.
          </Text>
        )}
      </section>
    </Box>
  );
};
