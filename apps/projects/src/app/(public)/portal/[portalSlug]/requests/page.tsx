import { PublicPortalRequestsPage } from "@/modules/public-portal";
import { parsePublicPortalFilters } from "@/modules/public-portal/query-params";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

export default async function PortalRequestsPage({
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
  const [portal, viewer] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, filters),
    getPublicPortalViewer(portalSlug),
  ]);

  return (
    <PublicPortalRequestsPage
      initialFilters={filters}
      portal={portal}
      viewer={viewer}
    />
  );
}
