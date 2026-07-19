import type { Metadata } from "next";
import type { PublicPortal } from "./types";

const FORTYONE_DOMAIN_SUFFIX = ".fortyone.app";
const FORTYONE_CLOUD_HOST = "cloud.fortyone.app";

const getFirstHeaderValue = (value?: string | null) =>
  value?.split(",")[0]?.trim();

export const getPublicPortalCanonicalUrl = ({
  forwardedHost,
  forwardedProtocol,
  host,
  portalSlug,
}: {
  forwardedHost?: string | null;
  forwardedProtocol?: string | null;
  host?: string | null;
  portalSlug: string;
}) => {
  const resolvedHost =
    getFirstHeaderValue(forwardedHost) ||
    getFirstHeaderValue(host) ||
    FORTYONE_CLOUD_HOST;
  const hostname = resolvedHost.split(":")[0] ?? resolvedHost;
  const requestedProtocol = getFirstHeaderValue(forwardedProtocol);
  let protocol = "https";
  if (requestedProtocol === "http" || requestedProtocol === "https") {
    protocol = requestedProtocol;
  } else if (hostname === "localhost") {
    protocol = "http";
  }
  const isWorkspaceSubdomain =
    hostname.endsWith(FORTYONE_DOMAIN_SUFFIX) &&
    hostname !== FORTYONE_CLOUD_HOST;
  const pathname = isWorkspaceSubdomain
    ? "/feedback"
    : `/portal/${encodeURIComponent(portalSlug)}/feedback`;

  return new URL(pathname, `${protocol}://${resolvedHost}`);
};

export const buildPublicPortalMetadata = (
  portal: PublicPortal,
  canonicalUrl: URL,
): Metadata => {
  const organizationName = portal.workspace.name;
  const title = `Send feedback to ${organizationName} | FortyOne`;
  const description = `Send requests and feedback directly to ${organizationName}. Browse existing ideas, vote on priorities, and follow public progress.`;

  return {
    alternates: {
      canonical: canonicalUrl,
    },
    applicationName: "FortyOne",
    description,
    keywords: [
      `${organizationName} feedback`,
      `${organizationName} requests`,
      "customer feedback",
      "feature requests",
      "public roadmap",
    ],
    openGraph: {
      description,
      siteName: "FortyOne",
      title,
      type: "website",
      url: canonicalUrl,
    },
    robots: {
      follow: true,
      index: true,
    },
    title,
    twitter: {
      card: "summary",
      description,
      title,
    },
  };
};
