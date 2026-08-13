import { notFound } from "next/navigation";
import { PublicPortalUpdatesPage } from "@/modules/public-portal";
import {
  getPublicPortalOrNotFound,
  getPublicPortalUpdates,
} from "@/modules/public-portal/query";
import { getPublicPortalParticipant } from "@/modules/public-portal/viewer";

export default async function PortalUpdatesPage({
  params,
}: {
  params: Promise<{ portalSlug: string }>;
}) {
  const { portalSlug } = await params;
  const [portal, participant, updatesPage] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug),
    getPublicPortalParticipant(portalSlug),
    getPublicPortalUpdates(portalSlug),
  ]);

  if (!portal.hasPublishedUpdates && updatesPage.updates.length === 0) {
    notFound();
  }

  return (
    <PublicPortalUpdatesPage
      participant={participant}
      portal={portal}
      updates={updatesPage.updates}
    />
  );
}
