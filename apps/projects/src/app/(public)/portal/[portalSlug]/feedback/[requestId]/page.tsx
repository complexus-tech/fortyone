import { notFound, redirect } from "next/navigation";
import { PublicPortalRequestDetailPage } from "@/modules/public-portal";
import {
  getPublicFeedbackCanonicalItemOrNotFound,
  getPublicPortalOrNotFound,
} from "@/modules/public-portal/query";
import { getPublicPortalParticipant } from "@/modules/public-portal/viewer";

type PageProps = {
  params: Promise<{ portalSlug: string; requestId: string }>;
};

export default async function PublicPortalFeedbackDetailRoute({
  params,
}: PageProps) {
  const { portalSlug, requestId } = await params;
  const canonical = await getPublicFeedbackCanonicalItemOrNotFound(
    portalSlug,
    requestId,
  );
  if (canonical.merged) {
    redirect(
      `/portal/${encodeURIComponent(portalSlug)}/feedback/${encodeURIComponent(canonical.itemSlug)}`,
    );
  }
  const [portal, participant] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, {
      itemId: canonical.itemId,
      pageSize: 1,
    }),
    getPublicPortalParticipant(portalSlug),
  ]);
  const request = portal.requests.find(
    (item) => item.id === canonical.itemId || item.slug === canonical.itemSlug,
  );

  if (!request) notFound();

  return (
    <PublicPortalRequestDetailPage
      participant={participant}
      portal={portal}
      request={request}
    />
  );
}
