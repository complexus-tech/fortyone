import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { features, getFeatureBySlug } from "@/lib/features";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";

export function generateStaticParams() {
  return features.map((feature) => ({ slug: feature.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const feature = getFeatureBySlug(slug);

  if (!feature) {
    return {};
  }

  const canonicalUrl = getCanonicalUrl(`/features/${feature.slug}`);

  return {
    title: feature.metaTitle,
    description: feature.metaDescription,
    alternates: {
      canonical: canonicalUrl,
    },
    openGraph: {
      type: "website",
      url: canonicalUrl,
      title: feature.metaTitle,
      description: feature.metaDescription,
      siteName: "FortyOne",
      images: [DEFAULT_SOCIAL_IMAGE],
    },
    twitter: {
      card: "summary_large_image",
      title: feature.metaTitle,
      description: feature.metaDescription,
      images: [DEFAULT_TWITTER_IMAGE],
    },
  };
}

export default async function FeaturePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const feature = getFeatureBySlug(slug);

  if (!feature) {
    return notFound();
  }

  return (
    <MarketingDetailPage
      basePath="features"
      breadcrumbLabel="Features"
      detail={feature}
      questionHeading={`Questions about ${feature.label.toLowerCase()}`}
    />
  );
}
