import { PublicPortalRequestsPage } from "@/modules/public-portal";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

export default async function PortalRequestsPage({
  params,
}: {
  params: Promise<{ portalSlug: string }>;
}) {
  const { portalSlug } = await params;
  const [portal, viewer] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug),
    getPublicPortalViewer(),
  ]);

  return <PublicPortalRequestsPage portal={portal} viewer={viewer} />;
}
