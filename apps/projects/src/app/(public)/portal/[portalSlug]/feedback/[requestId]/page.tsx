import { notFound } from "next/navigation";
import { PublicPortalRequestDetailPage } from "@/modules/public-portal";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

type PageProps = {
  params: Promise<{ portalSlug: string; requestId: string }>;
};

export default async function PublicPortalFeedbackDetailRoute({
  params,
}: PageProps) {
  const { portalSlug, requestId } = await params;
  const [portal, viewer] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, { pageSize: 1, search: requestId }),
    getPublicPortalViewer(portalSlug),
  ]);
  const request = portal.requests.find(
    (item) => item.slug === requestId || item.id === requestId,
  );

  if (!request) notFound();

  return (
    <PublicPortalRequestDetailPage
      portal={portal}
      request={request}
      viewer={viewer}
    />
  );
}
