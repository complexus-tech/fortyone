"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type {
  PublicPortalUpdate,
  PublicRequest,
} from "@/shared/feedback-widget/types";
import {
  exchangeWidgetIdentityAction,
  markWidgetFeedbackUpdatesSeenAction,
  revokeWidgetIdentityAction,
  toggleWidgetFeedbackVoteAction,
  type CreateWidgetFeedbackResult,
  type WidgetParticipantSession,
} from "./actions";
import {
  getTrustedWidgetOrigin,
  isFeedbackWidgetMessage,
  postFeedbackWidgetMessage,
} from "./protocol";
import type {
  FeedbackWidgetFrameProps,
  PendingIdentityAction,
  WidgetSubmissionIdentity,
} from "./components/types";
import { getFeedbackEmptyBody } from "./components/utils";
import { WidgetFrameView } from "./components/widget-frame-view";
import { useWidgetFeedbackData } from "./hooks/use-widget-feedback-data";
import {
  useWidgetFrameEmbedEvents,
  useWidgetTheme,
} from "./hooks/use-widget-frame-environment";

export const FeedbackWidgetFrame = ({
  initialTab,
  instanceId,
  mode,
  parentOrigin,
  portal,
  roadmap,
  roadmapPagination,
  theme,
  viewer,
}: FeedbackWidgetFrameProps) => {
  const rootRef = useRef<HTMLDivElement | null>(null);
  const identityExchangeRef = useRef(0);
  const [identity, setIdentity] = useState<WidgetSubmissionIdentity | null>(
    () => (viewer ? { kind: "account" } : null),
  );
  const activeIdentityRef = useRef(identity);
  const identityPendingRef = useRef(false);
  const updatesSeenTokenRef = useRef<string | null>(null);
  const [activeTab, setActiveTab] = useState(
    initialTab === "updates" && !portal.hasPublishedUpdates
      ? "home"
      : initialTab,
  );
  const {
    feedbackError,
    feedbackSort,
    feedbackStatus,
    homeRequests,
    isFeedbackLoading,
    loadMoreRoadmap,
    loadingRoadmapStatus,
    prependRequest,
    requests,
    roadmapError,
    roadmapItems,
    roadmapPageState,
    search,
    setFeedbackSort,
    setFeedbackStatus,
    setSearch,
    syncRequest: syncFeedbackRequest,
    visibleRoadmapCounts,
  } = useWidgetFeedbackData({
    initialRequests: portal.requests,
    initialRoadmap: roadmap,
    initialRoadmapPagination: roadmapPagination,
    portalSlug: portal.slug,
  });
  const [selectedRequest, setSelectedRequest] = useState<PublicRequest | null>(
    null,
  );
  const [selectedUpdate, setSelectedUpdate] =
    useState<PublicPortalUpdate | null>(null);
  const [isComposing, setIsComposing] = useState(false);
  const [composerIdentity, setComposerIdentity] =
    useState<WidgetSubmissionIdentity | null>(null);
  const [submissionSuccess, setSubmissionSuccess] =
    useState<CreateWidgetFeedbackResult | null>(null);
  const [identityError, setIdentityError] = useState("");
  const [isIdentityPending, setIsIdentityPending] = useState(false);
  const [unreadUpdateCount, setUnreadUpdateCount] = useState(0);
  const [pendingIdentityAction, setPendingIdentityAction] =
    useState<PendingIdentityAction | null>(null);
  const [votingRequestId, setVotingRequestId] = useState<string | null>(null);
  const trustedParentOrigin = getTrustedWidgetOrigin(parentOrigin);

  const revokeContributorIdentity = useCallback(
    (identityToRevoke: WidgetSubmissionIdentity | null) => {
      if (!identityToRevoke || identityToRevoke.kind === "account") return;
      void revokeWidgetIdentityAction({
        portalSlug: portal.slug,
        sessionToken: identityToRevoke.sessionToken,
      }).catch(() => undefined);
    },
    [portal.slug],
  );

  const beginIdentityTransition = useCallback(
    (pending: boolean) => {
      const previousIdentity = activeIdentityRef.current;
      const exchangeId = identityExchangeRef.current + 1;
      identityExchangeRef.current = exchangeId;
      identityPendingRef.current = pending;
      activeIdentityRef.current = null;
      updatesSeenTokenRef.current = null;
      setIdentity(null);
      setComposerIdentity(null);
      setIdentityError("");
      setIsIdentityPending(pending);
      setPendingIdentityAction(null);
      setUnreadUpdateCount(0);
      setVotingRequestId(null);
      revokeContributorIdentity(previousIdentity);
      return exchangeId;
    },
    [revokeContributorIdentity],
  );

  const activateIdentity = useCallback(
    (nextIdentity: WidgetSubmissionIdentity, nextUnreadUpdateCount = 0) => {
      identityPendingRef.current = false;
      activeIdentityRef.current = nextIdentity;
      updatesSeenTokenRef.current = null;
      setIdentity(nextIdentity);
      setComposerIdentity(nextIdentity);
      setIsIdentityPending(false);
      setUnreadUpdateCount(Math.max(0, nextUnreadUpdateCount));
    },
    [],
  );

  const canUseIdentity = useCallback(
    (candidate: WidgetSubmissionIdentity | null) =>
      !identityPendingRef.current && activeIdentityRef.current === candidate,
    [],
  );

  const syncRequest = useCallback(
    (updatedRequest: PublicRequest) => {
      syncFeedbackRequest(updatedRequest);
      setSelectedRequest((current) =>
        current?.id === updatedRequest.id ? updatedRequest : current,
      );
    },
    [syncFeedbackRequest],
  );

  const vote = useCallback(
    async (
      request: PublicRequest,
      activeIdentity: WidgetSubmissionIdentity,
      direction: -1 | 1,
    ) => {
      if (votingRequestId === request.id || !canUseIdentity(activeIdentity))
        return;
      const identityEpoch = identityExchangeRef.current;
      setVotingRequestId(request.id);
      const response = await toggleWidgetFeedbackVoteAction({
        itemId: request.id,
        participantKind: activeIdentity.kind,
        portalSlug: portal.slug,
        sessionToken:
          activeIdentity.kind === "account"
            ? undefined
            : activeIdentity.sessionToken,
        vote: direction,
      }).catch(() => null);
      if (
        identityExchangeRef.current !== identityEpoch ||
        !canUseIdentity(activeIdentity)
      )
        return;
      setVotingRequestId(null);
      if (!response?.data) return;
      syncRequest({
        ...request,
        viewerVote: response.data.vote,
        voteCount: response.data.voteCount,
      });
    },
    [canUseIdentity, portal.slug, syncRequest, votingRequestId],
  );

  const requestVote = useCallback(
    (request: PublicRequest, direction: -1 | 1 = 1) => {
      if (identityPendingRef.current) return;
      if (!identity) {
        setPendingIdentityAction({
          direction,
          identityEpoch: identityExchangeRef.current,
          requestId: request.id,
          type: "vote",
        });
        return;
      }
      void vote(request, identity, direction);
    },
    [identity, vote],
  );

  useWidgetTheme(theme);

  useEffect(() => {
    if (!trustedParentOrigin) return;
    postFeedbackWidgetMessage("ready", instanceId, trustedParentOrigin);
    const handleMessage = (event: MessageEvent) => {
      if (
        event.origin !== trustedParentOrigin ||
        event.source !== window.parent ||
        !isFeedbackWidgetMessage(event.data, instanceId)
      )
        return;

      if (event.data.event === "host-close") {
        setIsComposing(false);
        setComposerIdentity(null);
        setPendingIdentityAction(null);
        setSelectedRequest(null);
        setSelectedUpdate(null);
      }
      if (event.data.event === "host-identity-clear") {
        const requestIdValue = event.data.payload?.requestId;
        const requestId =
          typeof requestIdValue === "string" ||
          typeof requestIdValue === "number"
            ? String(requestIdValue)
            : undefined;
        beginIdentityTransition(false);
        postFeedbackWidgetMessage(
          "identity-cleared",
          instanceId,
          trustedParentOrigin,
          requestId ? { requestId } : undefined,
        );
      }
      if (event.data.event === "host-identify") {
        const requestIdValue = event.data.payload?.requestId;
        const requestId =
          typeof requestIdValue === "string" ||
          typeof requestIdValue === "number"
            ? String(requestIdValue)
            : undefined;
        const exchangeId = beginIdentityTransition(true);
        const assertion = event.data.payload?.assertion;
        if (typeof assertion !== "string" || assertion.length > 16384) {
          identityPendingRef.current = false;
          setIsIdentityPending(false);
          setIdentityError(
            "The product supplied an invalid customer identity.",
          );
          postFeedbackWidgetMessage(
            "identity-error",
            instanceId,
            trustedParentOrigin,
            {
              message: "invalid_assertion",
              ...(requestId ? { requestId } : {}),
            },
          );
          return;
        }
        void exchangeWidgetIdentityAction({
          assertion,
          parentOrigin: trustedParentOrigin,
          portalSlug: portal.slug,
        }).then(
          (response) => {
            if (identityExchangeRef.current !== exchangeId) {
              if (response.data) {
                revokeContributorIdentity({
                  kind: response.data.participant.kind,
                  sessionToken: response.data.session.token,
                });
              }
              return;
            }
            identityPendingRef.current = false;
            setIsIdentityPending(false);
            if (!response.data) {
              setIdentityError(
                response.error?.message ??
                  "Your customer identity could not be verified. Choose email verification or anonymous submission explicitly.",
              );
              postFeedbackWidgetMessage(
                "identity-error",
                instanceId,
                trustedParentOrigin,
                {
                  message: "exchange_failed",
                  ...(requestId ? { requestId } : {}),
                },
              );
              return;
            }
            activateIdentity(
              {
                kind: response.data.participant.kind,
                sessionToken: response.data.session.token,
              },
              response.data.participant.unreadUpdateCount,
            );
            postFeedbackWidgetMessage(
              "identity-ready",
              instanceId,
              trustedParentOrigin,
              {
                kind: response.data.participant.kind,
                ...(requestId ? { requestId } : {}),
              },
            );
          },
          () => {
            if (identityExchangeRef.current !== exchangeId) return;
            identityPendingRef.current = false;
            setIsIdentityPending(false);
            setIdentityError(
              "Your customer identity could not be verified. Choose email verification or anonymous submission explicitly.",
            );
            postFeedbackWidgetMessage(
              "identity-error",
              instanceId,
              trustedParentOrigin,
              {
                message: "exchange_failed",
                ...(requestId ? { requestId } : {}),
              },
            );
          },
        );
      }
    };
    window.addEventListener("message", handleMessage);
    return () => {
      window.removeEventListener("message", handleMessage);
    };
  }, [
    activateIdentity,
    beginIdentityTransition,
    instanceId,
    portal.slug,
    revokeContributorIdentity,
    trustedParentOrigin,
  ]);

  useEffect(() => {
    if (
      activeTab !== "updates" ||
      unreadUpdateCount <= 0 ||
      !identity ||
      identity.kind === "account" ||
      identityPendingRef.current ||
      updatesSeenTokenRef.current === identity.sessionToken
    )
      return;

    const identityEpoch = identityExchangeRef.current;
    const sessionToken = identity.sessionToken;
    updatesSeenTokenRef.current = sessionToken;
    void markWidgetFeedbackUpdatesSeenAction({
      portalSlug: portal.slug,
      sessionToken,
    }).then(
      (response) => {
        if (
          identityExchangeRef.current !== identityEpoch ||
          activeIdentityRef.current !== identity
        )
          return;
        if (!response.error) {
          setUnreadUpdateCount(0);
          return;
        }
        updatesSeenTokenRef.current = null;
      },
      () => {
        if (
          identityExchangeRef.current === identityEpoch &&
          activeIdentityRef.current === identity
        ) {
          updatesSeenTokenRef.current = null;
        }
      },
    );
  }, [activeTab, identity, portal.slug, unreadUpdateCount]);

  useWidgetFrameEmbedEvents(instanceId, mode, rootRef, trustedParentOrigin);

  const openComposer = () => {
    if (identityPendingRef.current) return;
    setComposerIdentity(activeIdentityRef.current);
    setIsComposing(true);
  };

  const handleVerified = (session: WidgetParticipantSession) => {
    const pendingAction = pendingIdentityAction;
    if (
      !pendingAction ||
      identityExchangeRef.current !== pendingAction.identityEpoch ||
      identityPendingRef.current
    ) {
      revokeContributorIdentity({
        kind: session.participant.kind,
        sessionToken: session.session.token,
      });
      return;
    }
    identityExchangeRef.current += 1;
    const verifiedIdentity: WidgetSubmissionIdentity = {
      kind: session.participant.kind,
      sessionToken: session.session.token,
    };
    activateIdentity(verifiedIdentity, session.participant.unreadUpdateCount);
    setPendingIdentityAction(null);
    if (pendingAction.type === "submit") {
      setComposerIdentity(verifiedIdentity);
      setIsComposing(true);
      return;
    }
    if (pendingAction.type === "vote") {
      const request =
        requests.find((item) => item.id === pendingAction.requestId) ??
        homeRequests.find((item) => item.id === pendingAction.requestId);
      if (request) {
        void vote(request, verifiedIdentity, pendingAction.direction);
      }
    }
  };

  const feedbackEmptyBody = getFeedbackEmptyBody(search, feedbackStatus);

  return (
    <WidgetFrameView
      activeTab={activeTab}
      feedback={{
        emptyBody: feedbackEmptyBody,
        error: feedbackError,
        isLoading: isFeedbackLoading,
        onRequestSelect: setSelectedRequest,
        onSearchChange: setSearch,
        onSortChange: setFeedbackSort,
        onStatusChange: setFeedbackStatus,
        onVote: requestVote,
        requests,
        search,
        sort: feedbackSort,
        status: feedbackStatus,
        votingRequestId,
        writeLocked: isIdentityPending,
      }}
      identity={{
        current: identity,
        error: identityError,
        isPending: isIdentityPending,
        unreadUpdateCount,
        use: canUseIdentity,
      }}
      mode={mode}
      onClose={() => {
        if (trustedParentOrigin) {
          postFeedbackWidgetMessage("close", instanceId, trustedParentOrigin);
        }
      }}
      onCommentCreated={(request, comment) => {
        syncRequest({
          ...request,
          commentCount: request.commentCount + 1,
          comments: [...request.comments, comment],
        });
      }}
      onComposerClosed={() => {
        setIsComposing(false);
        setComposerIdentity(null);
      }}
      onFeedbackCreated={(result) => {
        prependRequest(result.request);
        setIsComposing(false);
        setComposerIdentity(null);
        setSubmissionSuccess(result);
      }}
      onIdentityGateClosed={() => {
        setPendingIdentityAction(null);
      }}
      onIdentityVerified={handleVerified}
      onNavigate={(tab) => {
        setActiveTab(tab);
        setSelectedRequest(null);
        setSelectedUpdate(null);
      }}
      onOpenComposer={openComposer}
      onOpenExternal={() => {
        if (!trustedParentOrigin) return;
        postFeedbackWidgetMessage(
          "open-external",
          instanceId,
          trustedParentOrigin,
          {
            href: new URL(`/portal/${portal.slug}`, window.location.origin)
              .href,
          },
        );
      }}
      onRequestClose={() => {
        setSelectedRequest(null);
      }}
      onRequestSelect={setSelectedRequest}
      onRequireCommentIdentity={() => {
        setPendingIdentityAction({
          identityEpoch: identityExchangeRef.current,
          type: "comment",
        });
      }}
      onRequireComposerIdentity={() => {
        setPendingIdentityAction({
          identityEpoch: identityExchangeRef.current,
          type: "submit",
        });
      }}
      onSubmissionViewed={(result) => {
        setSelectedRequest(result.request);
        setSubmissionSuccess(null);
      }}
      onUpdateClose={() => {
        setSelectedUpdate(null);
      }}
      onUpdateSelect={setSelectedUpdate}
      onVote={requestVote}
      overlays={{
        composerIdentity,
        isComposing,
        isIdentityGateOpen: Boolean(pendingIdentityAction),
        selectedRequest,
        selectedUpdate,
        submissionSuccess,
      }}
      portal={portal}
      roadmap={{
        error: roadmapError,
        items: roadmapItems,
        loadingStatus: loadingRoadmapStatus,
        onLoadMore: (status) => {
          void loadMoreRoadmap(status);
        },
        onRequestSelect: setSelectedRequest,
        onVote: requestVote,
        pageState: roadmapPageState,
        visibleCounts: visibleRoadmapCounts,
        votingRequestId,
        writeLocked: isIdentityPending,
      }}
      rootRef={rootRef}
    />
  );
};
