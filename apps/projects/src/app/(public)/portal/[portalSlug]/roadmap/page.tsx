import { PublicPortalRoadmapPage } from "@/modules/public-portal";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

const ROADMAP_SERVER_CACHE_SECONDS = 5 * 60;

export default async function PortalRoadmapPage({
  params,
}: {
  params: Promise<{ portalSlug: string }>;
}) {
  const { portalSlug } = await params;
  const [portal, viewer] = await Promise.all([
    getPublicPortalOrNotFound(
      portalSlug,
      {},
      { revalidateSeconds: ROADMAP_SERVER_CACHE_SECONDS },
    ),
    getPublicPortalViewer(portalSlug),
  ]);

  return <PublicPortalRoadmapPage portal={portal} viewer={viewer} />;
}
