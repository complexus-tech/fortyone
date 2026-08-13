import { PublicPortalRequestsPage } from "@/modules/public-portal";
import {
  parsePublicPortalFilters,
  shouldOpenFeedbackComposer,
} from "@/modules/public-portal/query-params";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalParticipant } from "@/modules/public-portal/viewer";

type PageProps = {
  params: Promise<{ portalSlug: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export default async function PublicPortalFeedbackRoute({
  params,
  searchParams,
}: PageProps) {
  const [{ portalSlug }, resolvedSearchParams] = await Promise.all([
    params,
    searchParams,
  ]);
  const filters = parsePublicPortalFilters(resolvedSearchParams);
  const initialFeedbackComposerOpen =
    shouldOpenFeedbackComposer(resolvedSearchParams);
  const [portal, participant] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, filters),
    getPublicPortalParticipant(portalSlug),
  ]);

  return (
    <PublicPortalRequestsPage
      initialFeedbackComposerOpen={initialFeedbackComposerOpen}
      initialFilters={filters}
      participant={participant}
      portal={portal}
    />
  );
}
