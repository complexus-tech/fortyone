import type { ReactNode } from "react";
import type { Metadata } from "next";
import { headers } from "next/headers";
import {
  buildPublicPortalMetadata,
  getPublicPortalCanonicalUrl,
} from "@/modules/public-portal/metadata";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";

type LayoutProps = {
  children: ReactNode;
  params: Promise<{ portalSlug: string }>;
};

export const generateMetadata = async ({
  params,
}: LayoutProps): Promise<Metadata> => {
  const { portalSlug } = await params;
  const [portal, headerList] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, { sort: "top" }),
    headers(),
  ]);
  const canonicalUrl = getPublicPortalCanonicalUrl({
    forwardedHost: headerList.get("x-forwarded-host"),
    forwardedProtocol: headerList.get("x-forwarded-proto"),
    host: headerList.get("host"),
    portalSlug,
  });

  return buildPublicPortalMetadata(portal, canonicalUrl);
};

export default function PublicPortalLayout({ children }: LayoutProps) {
  return children;
}
