import { redirect } from "next/navigation";
import { getGlobalProfileHref } from "@/modules/public-portal/utils";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";
import { getLoginUrl } from "@/utils/callback-url";

export default async function PublicPortalViewerProfileRoute({
  params,
}: {
  params: Promise<{ portalSlug: string }>;
}) {
  const { portalSlug } = await params;
  const viewer = await getPublicPortalViewer(portalSlug);
  const profileHref = getGlobalProfileHref();

  if (!viewer) {
    redirect(getLoginUrl(profileHref));
  }

  redirect(profileHref);
}
