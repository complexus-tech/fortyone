import { PublicPortalRequestsPage } from "@/modules/public-portal";
import { parsePublicPortalFilters } from "@/modules/public-portal/query-params";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

type PageProps = {
  params: Promise<{ portalSlug: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export default async function PublicPortalFeedbackRoute({
  params,
  searchParams,
}: PageProps) {
  const [{ portalSlug }, resolvedSearchParams] = await Promise.all([
    params,
    searchParams,
  ]);
  const filters = parsePublicPortalFilters(resolvedSearchParams);
  const [portal, viewer] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, filters),
    getPublicPortalViewer(),
  ]);

  return (
    <PublicPortalRequestsPage
      initialFilters={filters}
      portal={portal}
      viewer={viewer}
    />
  );
}
