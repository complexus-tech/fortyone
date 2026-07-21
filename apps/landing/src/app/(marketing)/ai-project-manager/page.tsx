import type { Metadata } from "next";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { aiProjectManager } from "@/lib/ai-project-manager";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";

const canonicalUrl = getCanonicalUrl("/ai-project-manager");

export const metadata: Metadata = {
  title: aiProjectManager.metaTitle,
  description: aiProjectManager.metaDescription,
  alternates: {
    canonical: canonicalUrl,
  },
  openGraph: {
    title: aiProjectManager.metaTitle,
    description: aiProjectManager.metaDescription,
    siteName: "FortyOne",
    type: "website",
    url: canonicalUrl,
    images: [DEFAULT_SOCIAL_IMAGE],
  },
  twitter: {
    card: "summary_large_image",
    title: aiProjectManager.metaTitle,
    description: aiProjectManager.metaDescription,
    images: [DEFAULT_TWITTER_IMAGE],
  },
};

export default function AIProjectManagerPage() {
  return (
    <MarketingDetailPage
      basePath="ai-project-manager"
      breadcrumbLabel="AI project manager"
      canonicalPath="ai-project-manager"
      detail={aiProjectManager}
      questionHeading="Questions about AI project managers"
    />
  );
}
