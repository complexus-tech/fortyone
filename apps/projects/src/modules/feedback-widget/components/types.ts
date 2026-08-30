import type {
  PublicPortal,
  PublicPortalViewer,
  PublicRequest,
} from "@/shared/feedback-widget/types";
import type {
  FeedbackWidgetMode,
  FeedbackWidgetTab,
  FeedbackWidgetTheme,
} from "../protocol";

export type WidgetRoadmap = Record<
  "completed" | "in_progress" | "planned",
  PublicRequest[]
>;

export type WidgetRoadmapStatus = keyof WidgetRoadmap;

export type WidgetRoadmapPagination = Record<
  WidgetRoadmapStatus,
  { hasMore: boolean; nextPage: number }
>;

export type WidgetSubmissionIdentity =
  | { kind: "account" }
  | { kind: "external" | "verified_guest"; sessionToken: string };

export type PendingIdentityAction =
  | { identityEpoch: number; type: "comment" }
  | {
      direction: -1 | 1;
      identityEpoch: number;
      requestId: string;
      type: "vote";
    }
  | { identityEpoch: number; type: "submit" };

export type FeedbackWidgetFrameProps = {
  initialTab: FeedbackWidgetTab;
  instanceId: string;
  mode: FeedbackWidgetMode;
  parentOrigin: string;
  portal: PublicPortal;
  roadmap: WidgetRoadmap;
  roadmapPagination?: WidgetRoadmapPagination;
  theme: FeedbackWidgetTheme;
  viewer?: PublicPortalViewer | null;
};
