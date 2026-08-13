import { notFound, redirect } from "next/navigation";
import { PublicPortalRequestDetailPage } from "@/modules/public-portal";
import {
  getPublicFeedbackCanonicalItemOrNotFound,
  getPublicPortalOrNotFound,
} from "@/modules/public-portal/query";
import { getPublicPortalParticipant } from "@/modules/public-portal/viewer";

export default async function PortalRequestDetailPage({
  params,
}: {
  params: Promise<{ portalSlug: string; requestId: string }>;
}) {
  const { portalSlug, requestId } = await params;
  const canonical = await getPublicFeedbackCanonicalItemOrNotFound(
    portalSlug,
    requestId,
  );
  if (canonical.merged) {
    redirect(
      `/portal/${encodeURIComponent(portalSlug)}/requests/${encodeURIComponent(canonical.itemSlug)}`,
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
    (publicRequest) =>
      publicRequest.id === canonical.itemId ||
      publicRequest.slug === canonical.itemSlug,
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
