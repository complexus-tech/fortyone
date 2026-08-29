import { PublicPortalUpdateDetailPage } from "@/modules/public-portal";
import {
  getPublicPortalOrNotFound,
  getPublicPortalUpdateOrNotFound,
} from "@/modules/public-portal/query";
import { getPublicPortalParticipant } from "@/modules/public-portal/viewer";

export default async function PortalUpdateDetailPage({
  params,
}: {
  params: Promise<{ portalSlug: string; updateSlug: string }>;
}) {
  const { portalSlug, updateSlug } = await params;
  const [portal, participant, update] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, {}, { revalidateSeconds: 300 }),
    getPublicPortalParticipant(portalSlug),
    getPublicPortalUpdateOrNotFound(portalSlug, updateSlug),
  ]);

  return (
    <PublicPortalUpdateDetailPage
      participant={participant}
      portal={portal}
      update={update}
    />
  );
}
