import { headers } from "next/headers";
import { notFound } from "next/navigation";
import { PublicPortalRequestDetailPage } from "@/modules/public-portal";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";
import { getTeams } from "@/lib/queries/get-teams";

type PageProps = {
  params: Promise<{ portalSlug: string; requestId: string }>;
};

const DOMAIN_SUFFIX = ".fortyone.app";

const getWorkspaceSlug = async (portalSlug: string) => {
  const host = (await headers()).get("host")?.split(":")[0] ?? "";

  if (host.endsWith(DOMAIN_SUFFIX)) {
    const subdomain = host.replace(DOMAIN_SUFFIX, "");
    if (subdomain && subdomain !== "cloud") return subdomain;
  }

  return portalSlug;
};

export default async function PublicPortalFeedbackDetailRoute({
  params,
}: PageProps) {
  const { portalSlug, requestId } = await params;
  const workspaceSlug = await getWorkspaceSlug(portalSlug);
  const [portal, viewer, teams] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, { pageSize: 1, search: requestId }),
    getPublicPortalViewer(),
    getTeams(workspaceSlug),
  ]);
  const request = portal.requests.find(
    (item) => item.slug === requestId || item.id === requestId,
  );

  if (!request) notFound();

  return (
    <PublicPortalRequestDetailPage
      portal={portal}
      request={request}
      teams={teams}
      viewer={viewer}
    />
  );
}
