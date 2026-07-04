import { PublicPortalRequestsPage } from "@/modules/public-portal";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

type PageProps = {
  params: Promise<{ portalSlug: string }>;
};

export default async function PublicPortalFeedbackRoute({ params }: PageProps) {
  const { portalSlug } = await params;
  const [portal, viewer] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug),
    getPublicPortalViewer(),
  ]);

  return <PublicPortalRequestsPage portal={portal} viewer={viewer} />;
}
