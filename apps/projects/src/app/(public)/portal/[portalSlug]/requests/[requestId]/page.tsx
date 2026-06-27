import { PublicPortalRequestDetailPage } from "@/modules/public-portal";
import { getPublicPortal } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

export default async function PortalRequestDetailPage({
  params,
}: {
  params: Promise<{ portalSlug: string; requestId: string }>;
}) {
  const { portalSlug, requestId } = await params;
  const [portal, viewer] = await Promise.all([
    getPublicPortal(portalSlug),
    getPublicPortalViewer(),
  ]);
  const request =
    portal.requests.find(
      (publicRequest) =>
        publicRequest.slug === requestId || publicRequest.id === requestId,
    ) ?? portal.requests[0];

  return (
    <PublicPortalRequestDetailPage
      portal={portal}
      request={request}
      viewer={viewer}
    />
  );
}
