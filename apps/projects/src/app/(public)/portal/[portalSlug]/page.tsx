import { PublicPortalRequestsPage } from "@/modules/public-portal";
import { parsePublicPortalFilters } from "@/modules/public-portal/query-params";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalParticipant } from "@/modules/public-portal/viewer";

export default async function PortalPage({
  params,
  searchParams,
}: {
  params: Promise<{ portalSlug: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const [{ portalSlug }, resolvedSearchParams] = await Promise.all([
    params,
    searchParams,
  ]);
  const filters = parsePublicPortalFilters(resolvedSearchParams);
  const [portal, participant] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, filters),
    getPublicPortalParticipant(portalSlug),
  ]);

  return (
    <PublicPortalRequestsPage
      initialFilters={filters}
      participant={participant}
      portal={portal}
    />
  );
}
