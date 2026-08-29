import { PublicPortalRoadmapPage } from "@/modules/public-portal";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalParticipant } from "@/modules/public-portal/viewer";

const ROADMAP_SERVER_CACHE_SECONDS = 5 * 60;

export default async function PortalRoadmapPage({
  params,
}: {
  params: Promise<{ portalSlug: string }>;
}) {
  const { portalSlug } = await params;
  const [portal, participant] = await Promise.all([
    getPublicPortalOrNotFound(
      portalSlug,
      {},
      { revalidateSeconds: ROADMAP_SERVER_CACHE_SECONDS },
    ),
    getPublicPortalParticipant(portalSlug),
  ]);

  return <PublicPortalRoadmapPage participant={participant} portal={portal} />;
}
