import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { FeedbackWidgetFrame } from "@/modules/feedback-widget/widget-frame";
import {
  getTrustedWidgetOrigin,
  type FeedbackWidgetMode,
  type FeedbackWidgetTab,
  type FeedbackWidgetTheme,
} from "@/modules/feedback-widget/protocol";
import {
  getPublicFeedbackPortalOrNotFound,
  getPublicPortalOrNotFound,
  getPublicPortalUpdates,
} from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

export const metadata: Metadata = {
  robots: {
    follow: false,
    index: false,
  },
  title: "Feedback",
};

const first = (value: string | string[] | undefined) =>
  Array.isArray(value) ? value[0] : value;

const parseTab = (value?: string): FeedbackWidgetTab =>
  value === "roadmap" || value === "updates" ? value : "feedback";

const parseMode = (value?: string): FeedbackWidgetMode =>
  value === "custom" || value === "inline" ? value : "bubble";

const parseTheme = (value?: string): FeedbackWidgetTheme =>
  value === "light" || value === "dark" ? value : "auto";

export default async function FeedbackWidgetPage({
  params,
  searchParams,
}: {
  params: Promise<{ portalSlug: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const [{ portalSlug }, query] = await Promise.all([params, searchParams]);
  const instanceId = first(query.instance)?.trim();
  const parentOrigin = getTrustedWidgetOrigin(first(query.parentOrigin));
  if (!instanceId || instanceId.length > 160 || !parentOrigin) notFound();

  const [portal, planned, inProgress, completed, viewer, updatesPage] =
    await Promise.all([
      getPublicPortalOrNotFound(
        portalSlug,
        {
          pageSize: 20,
          sort: "top",
          status: "active",
          view: "summary",
        },
        { revalidateSeconds: 30 },
      ),
      getPublicFeedbackPortalOrNotFound(
        portalSlug,
        { pageSize: 20, sort: "newest", status: "planned", view: "summary" },
        { revalidateSeconds: 5 * 60 },
      ),
      getPublicFeedbackPortalOrNotFound(
        portalSlug,
        {
          pageSize: 20,
          sort: "newest",
          status: "in_progress",
          view: "summary",
        },
        { revalidateSeconds: 5 * 60 },
      ),
      getPublicFeedbackPortalOrNotFound(
        portalSlug,
        { pageSize: 20, sort: "newest", status: "completed", view: "summary" },
        { revalidateSeconds: 5 * 60 },
      ),
      getPublicPortalViewer(portalSlug),
      getPublicPortalUpdates(portalSlug, 1, 20).catch(() => ({
        hasMore: false,
        unreadCount: 0,
        updates: [],
      })),
    ]);

  return (
    <FeedbackWidgetFrame
      initialTab={parseTab(first(query.tab))}
      instanceId={instanceId}
      mode={parseMode(first(query.mode))}
      parentOrigin={parentOrigin}
      portal={{ ...portal, updates: updatesPage.updates }}
      roadmap={{
        completed: completed.requests,
        in_progress: inProgress.requests,
        planned: planned.requests,
      }}
      theme={parseTheme(first(query.theme))}
      viewer={viewer}
    />
  );
}
