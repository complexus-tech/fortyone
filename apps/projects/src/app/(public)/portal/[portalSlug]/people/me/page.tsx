import { redirect } from "next/navigation";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import {
  getAuthorPath,
  getPortalPath,
  getViewerProfileCallbackUrl,
} from "@/modules/public-portal/utils";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";
import { getLoginUrl } from "@/utils/callback-url";

export default async function PublicPortalViewerProfileRoute({
  params,
}: {
  params: Promise<{ portalSlug: string }>;
}) {
  const { portalSlug } = await params;
  const [portal, viewer] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug),
    getPublicPortalViewer(portalSlug),
  ]);

  if (!viewer) {
    redirect(getLoginUrl(getViewerProfileCallbackUrl(portal)));
  }

  redirect(
    getAuthorPath(portal, viewer.id) ?? getPortalPath(portal, "feedback"),
  );
}
