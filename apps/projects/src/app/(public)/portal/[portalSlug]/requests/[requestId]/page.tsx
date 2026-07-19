import { notFound } from "next/navigation";
import { PublicPortalRequestDetailPage } from "@/modules/public-portal";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

export default async function PortalRequestDetailPage({
  params,
}: {
  params: Promise<{ portalSlug: string; requestId: string }>;
}) {
  const { portalSlug, requestId } = await params;
  const [portal, viewer] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, { pageSize: 1, search: requestId }),
    getPublicPortalViewer(portalSlug),
  ]);
  const request = portal.requests.find(
    (publicRequest) =>
      publicRequest.slug === requestId || publicRequest.id === requestId,
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
