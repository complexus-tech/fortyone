import type { RefObject } from "react";
import {
  BellIcon,
  CloseIcon,
  EditIcon,
  HomeIcon,
  RequestsIcon,
  RoadmapIcon,
  UpdatesIcon,
} from "icons";
import { Avatar, Box, Button, Flex, Text } from "ui";
import { cn, getReadableTextColor } from "lib";
import type {
  PublicFeedbackListStatus,
  PublicPortal,
  PublicPortalSort,
  PublicPortalUpdate,
  PublicRequest,
  PublicRequestComment,
} from "@/shared/feedback-widget/types";
import type {
  CreateWidgetFeedbackResult,
  WidgetParticipantSession,
} from "../actions";
import type { FeedbackWidgetMode, FeedbackWidgetTab } from "../protocol";
import { FeedbackComposer, SubmissionSuccess } from "./feedback-composer";
import { WidgetHome, RoadmapGroupHeader } from "./home";
import { IdentityGate } from "./identity-gate";
import { RequestDetail } from "./request-detail";
import type {
  WidgetRoadmap,
  WidgetRoadmapPagination,
  WidgetRoadmapStatus,
  WidgetSubmissionIdentity,
} from "./types";
import { roadmapSections, statusAccent } from "./utils";
import { UpdateDetail, UpdatesList } from "./updates";
import {
  EmptyState,
  FeedbackRow,
  UnreadBadge,
  VoteButton,
  WidgetFeedbackToolbar,
  WidgetIconButton,
} from "./widget-ui";

const tabs = [
  { icon: HomeIcon, label: "Home", value: "home" },
  { icon: RequestsIcon, label: "Feedback", value: "feedback" },
  { icon: RoadmapIcon, label: "Roadmap", value: "roadmap" },
  { icon: UpdatesIcon, label: "Updates", value: "updates" },
] satisfies {
  icon: typeof RequestsIcon;
  label: string;
  value: FeedbackWidgetTab;
}[];

const BottomNavigation = ({
  activeTab,
  onSelect,
  showUpdates,
  unreadUpdateCount,
}: {
  activeTab: FeedbackWidgetTab;
  onSelect: (tab: FeedbackWidgetTab) => void;
  showUpdates: boolean;
  unreadUpdateCount: number;
}) => (
  <nav
    aria-label="Feedback sections"
    className={cn(
      "border-border/70 bg-background/85 supports-[backdrop-filter]:bg-background/75 grid shrink-0 border-t px-2 py-2 backdrop-blur-xl",
      showUpdates ? "grid-cols-4" : "grid-cols-3",
    )}
  >
    {tabs
      .filter((tab) => tab.value !== "updates" || showUpdates)
      .map((tab) => {
        const Icon = tab.icon;
        const active = activeTab === tab.value;
        const unreadCount = tab.value === "updates" ? unreadUpdateCount : 0;
        return (
          <button
            aria-current={active ? "page" : undefined}
            aria-label={
              unreadCount > 0
                ? `${tab.label}, ${unreadCount} unread ${unreadCount === 1 ? "update" : "updates"}`
                : tab.label
            }
            className={cn(
              "text-text-muted hover:text-foreground focus-visible:ring-ring flex h-12 flex-col items-center justify-center gap-1 rounded-lg text-[12px] font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none",
              { "text-foreground": active },
            )}
            key={tab.value}
            onClick={() => {
              onSelect(tab.value);
            }}
            type="button"
          >
            <span className="relative">
              <Icon className="h-[18px] text-current" />
              <UnreadBadge count={unreadCount} />
            </span>
            {tab.label}
          </button>
        );
      })}
  </nav>
);

type FeedbackTabProps = {
  emptyBody: string;
  error: string;
  isLoading: boolean;
  onRequestSelect: (request: PublicRequest) => void;
  onSearchChange: (value: string) => void;
  onSortChange: (value: PublicPortalSort) => void;
  onStatusChange: (value: PublicFeedbackListStatus) => void;
  onVote: (request: PublicRequest) => void;
  requests: PublicRequest[];
  search: string;
  sort: PublicPortalSort;
  status: PublicFeedbackListStatus;
  votingRequestId: string | null;
  writeLocked: boolean;
};

const FeedbackTab = ({
  emptyBody,
  error,
  isLoading,
  onRequestSelect,
  onSearchChange,
  onSortChange,
  onStatusChange,
  onVote,
  requests,
  search,
  sort,
  status,
  votingRequestId,
  writeLocked,
}: FeedbackTabProps) => (
  <Box>
    <WidgetFeedbackToolbar
      isLoading={isLoading}
      onSearchChange={onSearchChange}
      onSortChange={onSortChange}
      onStatusChange={onStatusChange}
      search={search}
      sort={sort}
      status={status}
    />
    {error ? (
      <Text
        aria-live="polite"
        className="border-border/60 border-b px-5 py-2 text-[11px] text-red-600 dark:text-red-400"
      >
        {error}
      </Text>
    ) : null}
    <Box
      aria-busy={isLoading}
      className={cn("transition-opacity", { "opacity-60": isLoading })}
    >
      {requests.length > 0 ? (
        requests.map((request) => (
          <FeedbackRow
            isVoting={votingRequestId === request.id}
            isWriteLocked={writeLocked}
            key={request.id}
            onOpen={() => {
              onRequestSelect(request);
            }}
            onVote={() => {
              onVote(request);
            }}
            request={request}
          />
        ))
      ) : (
        <EmptyState
          body={emptyBody}
          icon={RequestsIcon}
          title={search ? "No matching feedback" : "No feedback yet"}
        />
      )}
    </Box>
  </Box>
);

type RoadmapTabProps = {
  error: string;
  items: WidgetRoadmap;
  loadingStatus: WidgetRoadmapStatus | null;
  onLoadMore: (status: WidgetRoadmapStatus) => void;
  onRequestSelect: (request: PublicRequest) => void;
  onVote: (request: PublicRequest) => void;
  pageState: WidgetRoadmapPagination;
  visibleCounts: Record<WidgetRoadmapStatus, number>;
  votingRequestId: string | null;
  writeLocked: boolean;
};

const RoadmapTab = ({
  error,
  items,
  loadingStatus,
  onLoadMore,
  onRequestSelect,
  onVote,
  pageState,
  visibleCounts,
  votingRequestId,
  writeLocked,
}: RoadmapTabProps) => (
  <Box className="px-5 py-6">
    {error ? (
      <Text
        aria-live="polite"
        className="text-[11px] text-red-600 dark:text-red-400"
      >
        {error}
      </Text>
    ) : null}
    <Box className="border-border/70 ml-1.5 border-l border-dashed">
      {roadmapSections.map((section) => {
        const sectionItems = items[section.status];
        const visibleItems = sectionItems.slice(
          0,
          visibleCounts[section.status],
        );
        const hasMore =
          visibleItems.length < sectionItems.length ||
          pageState[section.status].hasMore;
        return (
          <Box key={section.status}>
            <RoadmapGroupHeader
              count={sectionItems.length}
              label={section.label}
              status={section.status}
            />
            {sectionItems.length > 0 ? (
              <Box>
                {visibleItems.map((request) => (
                  <div
                    className="hover:bg-state-hover/35 relative grid w-full grid-cols-[minmax(0,1fr)_auto] items-start gap-3 py-3 pr-1 pl-6 transition-colors"
                    key={request.id}
                  >
                    <span
                      className={cn(
                        "ring-background absolute top-5 -left-[5px] size-2.5 rounded-sm ring-4",
                        statusAccent(section.status),
                      )}
                    />
                    <button
                      className="focus-visible:ring-ring min-w-0 text-left focus-visible:ring-2 focus-visible:outline-none"
                      onClick={() => {
                        onRequestSelect(request);
                      }}
                      type="button"
                    >
                      <Box className="min-w-0">
                        <Text
                          className="line-clamp-1 text-[13px]"
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
                        <Flex align="center" className="mt-2 min-w-0 gap-1.5">
                          <Avatar
                            name={request.authorName || "Anonymous"}
                            size="xs"
                            src={request.authorAvatar}
                          />
                          <Text
                            className="truncate text-[12px] font-medium"
                            color="muted"
                          >
                            {request.authorName || "Anonymous"}
                          </Text>
                        </Flex>
                      </Box>
                    </button>
                    <VoteButton
                      disabled={writeLocked}
                      isPending={votingRequestId === request.id}
                      onClick={() => {
                        onVote(request);
                      }}
                      request={request}
                    />
                  </div>
                ))}
                {hasMore ? (
                  <button
                    className="text-text-muted hover:text-foreground focus-visible:ring-ring inline-flex h-9 items-center rounded-lg pr-2 pl-6 text-[12px] font-semibold transition-colors focus-visible:ring-2 focus-visible:outline-none disabled:cursor-wait disabled:opacity-60"
                    disabled={loadingStatus !== null}
                    onClick={() => {
                      onLoadMore(section.status);
                    }}
                    type="button"
                  >
                    {loadingStatus === section.status
                      ? "Loading…"
                      : "Show more"}
                  </button>
                ) : null}
              </Box>
            ) : (
              <Text className="py-3 pr-1 pl-6 text-[12px]" color="muted">
                Nothing here yet
              </Text>
            )}
          </Box>
        );
      })}
    </Box>
  </Box>
);

export const WidgetFrameView = ({
  activeTab,
  feedback,
  identity,
  mode,
  onClose,
  onCommentCreated,
  onComposerClosed,
  onFeedbackCreated,
  onIdentityGateClosed,
  onIdentityVerified,
  onNavigate,
  onOpenComposer,
  onOpenExternal,
  onRequireCommentIdentity,
  onRequireComposerIdentity,
  onRequestClose,
  onRequestSelect,
  onSubmissionViewed,
  onUpdateSelect,
  onUpdateClose,
  onVote,
  overlays,
  portal,
  roadmap,
  rootRef,
}: {
  activeTab: FeedbackWidgetTab;
  feedback: FeedbackTabProps;
  identity: {
    current: WidgetSubmissionIdentity | null;
    error: string;
    isPending: boolean;
    unreadUpdateCount: number;
    use: (identity: WidgetSubmissionIdentity | null) => boolean;
  };
  mode: FeedbackWidgetMode;
  onClose: () => void;
  onCommentCreated: (
    request: PublicRequest,
    comment: PublicRequestComment,
  ) => void;
  onComposerClosed: () => void;
  onFeedbackCreated: (result: CreateWidgetFeedbackResult) => void;
  onIdentityGateClosed: () => void;
  onIdentityVerified: (session: WidgetParticipantSession) => void;
  onNavigate: (tab: FeedbackWidgetTab) => void;
  onOpenComposer: () => void;
  onOpenExternal: () => void;
  onRequireCommentIdentity: () => void;
  onRequireComposerIdentity: () => void;
  onRequestClose: () => void;
  onRequestSelect: (request: PublicRequest) => void;
  onSubmissionViewed: (result: CreateWidgetFeedbackResult) => void;
  onUpdateSelect: (update: PublicPortalUpdate) => void;
  onUpdateClose: () => void;
  onVote: (request: PublicRequest, direction?: -1 | 1) => void;
  overlays: {
    composerIdentity: WidgetSubmissionIdentity | null;
    isComposing: boolean;
    isIdentityGateOpen: boolean;
    selectedRequest: PublicRequest | null;
    selectedUpdate: PublicPortalUpdate | null;
    submissionSuccess: CreateWidgetFeedbackResult | null;
  };
  portal: PublicPortal;
  roadmap: RoadmapTabProps;
  rootRef: RefObject<HTMLDivElement | null>;
}) => {
  const {
    current: currentIdentity,
    error: identityError,
    isPending,
  } = identity;
  const { selectedRequest, selectedUpdate, submissionSuccess } = overlays;

  return (
    <div
      className="bg-background/90 supports-[backdrop-filter]:bg-background/80 text-foreground relative flex h-dvh min-h-0 w-full flex-col overflow-hidden antialiased backdrop-blur-xl"
      ref={rootRef}
    >
      <Flex
        align="center"
        className="bg-background/85 supports-[backdrop-filter]:bg-background/75 relative z-10 h-16 shrink-0 gap-3 px-5 backdrop-blur-xl"
      >
        <Avatar
          className="!size-8 text-[12px] font-bold"
          name={portal.workspace.name}
          rounded="lg"
          size="sm"
          src={portal.workspace.avatarUrl}
          style={{
            backgroundColor: portal.workspace.color,
            color: getReadableTextColor(portal.workspace.color),
          }}
        />
        <Text
          className="min-w-0 flex-1 truncate text-[15px]"
          fontWeight="semibold"
        >
          {portal.workspace.name}
        </Text>
        <Button
          className="h-9 shrink-0 gap-2 rounded-xl px-3.5 text-[12px] shadow-sm"
          color="invert"
          disabled={isPending}
          leftIcon={<EditIcon className="h-3.5 text-current" />}
          onClick={onOpenComposer}
          size="sm"
        >
          Add feedback
        </Button>
        {currentIdentity && portal.hasPublishedUpdates ? (
          <WidgetIconButton
            aria-label={
              identity.unreadUpdateCount > 0
                ? `View feedback updates, ${identity.unreadUpdateCount} unread ${identity.unreadUpdateCount === 1 ? "update" : "updates"}`
                : "View feedback updates"
            }
            className="relative hidden min-[380px]:inline-flex"
            onClick={() => {
              onNavigate("updates");
            }}
          >
            <BellIcon className="h-[18px]" />
            <UnreadBadge count={identity.unreadUpdateCount} />
          </WidgetIconButton>
        ) : null}
        {mode !== "inline" ? (
          <WidgetIconButton
            aria-label="Close feedback widget"
            onClick={onClose}
          >
            <CloseIcon className="h-5" />
          </WidgetIconButton>
        ) : null}
      </Flex>
      {isPending ? (
        <Text
          className="border-border/70 border-t px-5 py-2 text-[11px]"
          color="muted"
        >
          Verifying your customer identity…
        </Text>
      ) : null}
      {identityError ? (
        <Text className="border-border/70 border-t px-5 py-2 text-[11px] text-red-600 dark:text-red-400">
          {identityError}
        </Text>
      ) : null}

      <Box className="relative min-h-0 flex-1 overflow-y-auto">
        {activeTab === "home" ? (
          <WidgetHome
            feedback={feedback.requests}
            isWriteLocked={isPending}
            latestUpdate={
              portal.hasPublishedUpdates ? portal.updates[0] : undefined
            }
            onOpenFeedback={() => {
              onNavigate("feedback");
            }}
            onOpenRequest={onRequestSelect}
            onOpenRoadmap={() => {
              onNavigate("roadmap");
            }}
            onOpenUpdate={onUpdateSelect}
            onShareFeedback={onOpenComposer}
            onVote={(request) => {
              onVote(request);
            }}
            roadmap={roadmap.items}
            votingRequestId={feedback.votingRequestId}
          />
        ) : null}

        {activeTab === "feedback" ? <FeedbackTab {...feedback} /> : null}

        {activeTab === "roadmap" ? <RoadmapTab {...roadmap} /> : null}

        {activeTab === "updates" ? (
          <UpdatesList onOpen={onUpdateSelect} updates={portal.updates} />
        ) : null}
      </Box>

      <BottomNavigation
        activeTab={activeTab}
        onSelect={onNavigate}
        showUpdates={portal.hasPublishedUpdates}
        unreadUpdateCount={identity.unreadUpdateCount}
      />
      <button
        className="border-border/70 text-text-muted hover:text-foreground bg-background/85 supports-[backdrop-filter]:bg-background/75 focus-visible:ring-ring flex h-12 shrink-0 items-center justify-center border-t px-5 py-3 text-[12px] backdrop-blur-xl transition-colors focus-visible:ring-2 focus-visible:outline-none"
        onClick={onOpenExternal}
        type="button"
      >
        Powered by{" "}
        <span className="ml-1 font-semibold text-[var(--color-foreground)]">
          FortyOne
        </span>
        <span aria-hidden="true" className="ml-1">
          ↗
        </span>
      </button>

      {selectedRequest ? (
        <RequestDetail
          canUseIdentity={identity.use}
          identity={currentIdentity}
          isVoting={feedback.votingRequestId === selectedRequest.id}
          isWriteLocked={isPending}
          onBack={onRequestClose}
          onCommentCreated={(comment) => {
            onCommentCreated(selectedRequest, comment);
          }}
          onRequireIdentity={onRequireCommentIdentity}
          onVote={(direction) => {
            onVote(selectedRequest, direction);
          }}
          portal={portal}
          request={selectedRequest}
        />
      ) : null}
      {selectedUpdate ? (
        <UpdateDetail onBack={onUpdateClose} update={selectedUpdate} />
      ) : null}
      {overlays.isComposing ? (
        <FeedbackComposer
          canUseIdentity={identity.use}
          identity={overlays.composerIdentity}
          isWriteLocked={isPending}
          onBack={onComposerClosed}
          onCreated={onFeedbackCreated}
          onRequireIdentity={onRequireComposerIdentity}
          portal={portal}
        />
      ) : null}
      {submissionSuccess ? (
        <SubmissionSuccess
          onView={() => {
            onSubmissionViewed(submissionSuccess);
          }}
          request={submissionSuccess.request}
        />
      ) : null}
      {overlays.isIdentityGateOpen ? (
        <IdentityGate
          onBack={onIdentityGateClosed}
          onVerified={onIdentityVerified}
          portal={portal}
        />
      ) : null}
    </div>
  );
};
