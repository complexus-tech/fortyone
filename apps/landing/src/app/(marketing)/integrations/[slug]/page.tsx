import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { getIntegrationBySlug, integrations } from "@/lib/integrations";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";

export function generateStaticParams() {
  return integrations.map((integration) => ({ slug: integration.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const integration = getIntegrationBySlug(slug);

  if (!integration) {
    return {};
  }

  const canonicalUrl = getCanonicalUrl(`/integrations/${integration.slug}`);

  return {
    title: integration.metaTitle,
    description: integration.metaDescription,
    alternates: {
      canonical: canonicalUrl,
    },
    openGraph: {
      type: "website",
      url: canonicalUrl,
      title: integration.metaTitle,
      description: integration.metaDescription,
      siteName: "FortyOne",
      images: [DEFAULT_SOCIAL_IMAGE],
    },
    twitter: {
      card: "summary_large_image",
      title: integration.metaTitle,
      description: integration.metaDescription,
      images: [DEFAULT_TWITTER_IMAGE],
    },
  };
}

export default async function IntegrationPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const integration = getIntegrationBySlug(slug);

  if (!integration) {
    return notFound();
  }

  return (
    <MarketingDetailPage
      basePath="integrations"
      breadcrumbLabel="Integrations"
      detail={integration}
      questionHeading={`Questions about the ${integration.label} integration`}
    />
  );
}
