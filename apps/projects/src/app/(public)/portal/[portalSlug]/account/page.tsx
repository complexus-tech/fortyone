import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { PublicPortalAccountPage } from "@/modules/public-portal/account-page";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPortalLoginUrl } from "@/modules/public-portal/utils";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";
import { getProfile } from "@/lib/queries/profile";

export const metadata: Metadata = {
  title: "Account settings",
  robots: {
    follow: false,
    index: false,
  },
};

export default async function PublicPortalAccountRoute({
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
    redirect(getPortalLoginUrl(portal, "account"));
  }

  const profile = await getProfile();

  return (
    <PublicPortalAccountPage
      portal={portal}
      profile={profile}
      viewer={viewer}
    />
  );
}
