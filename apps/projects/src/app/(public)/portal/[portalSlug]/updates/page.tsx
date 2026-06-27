import { PublicPortalUpdatesPage } from "@/modules/public-portal";
import { getPublicPortal } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

export default async function PortalUpdatesPage({
  params,
}: {
  params: Promise<{ portalSlug: string }>;
}) {
  const { portalSlug } = await params;
  const [portal, viewer] = await Promise.all([
    getPublicPortal(portalSlug),
    getPublicPortalViewer(),
  ]);

  return <PublicPortalUpdatesPage portal={portal} viewer={viewer} />;
}
