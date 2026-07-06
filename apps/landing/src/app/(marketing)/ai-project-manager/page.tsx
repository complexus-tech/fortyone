import type { Metadata } from "next";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { aiProjectManager } from "@/lib/ai-project-manager";

export const metadata: Metadata = {
  title: aiProjectManager.metaTitle,
  description: aiProjectManager.metaDescription,
  alternates: {
    canonical: "https://www.fortyone.app/ai-project-manager",
  },
  openGraph: {
    title: aiProjectManager.metaTitle,
    description: aiProjectManager.metaDescription,
    siteName: "FortyOne",
    type: "website",
    url: "https://www.fortyone.app/ai-project-manager",
  },
  twitter: {
    card: "summary_large_image",
    title: aiProjectManager.metaTitle,
    description: aiProjectManager.metaDescription,
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
