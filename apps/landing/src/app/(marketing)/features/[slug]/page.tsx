import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { features, getFeatureBySlug } from "@/lib/features";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";
import { AiPlanningPage } from "@/modules/features/ai-planning-page";
import { FeatureLandingPage } from "@/modules/features/feature-landing-page";
import { getFeaturePageConfig } from "@/modules/features/feature-page-config";

const AI_PLANNING_KEYWORDS = [
  "AI project planning",
  "AI planning software",
  "AI capacity planning",
  "project planning assistant",
  "AI project management",
  "project risk detection",
  "team workload planning",
];

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
  const featurePageConfig = getFeaturePageConfig(feature.slug);

  return {
    title: feature.metaTitle,
    description: feature.metaDescription,
    keywords:
      featurePageConfig?.keywords ??
      (feature.slug === "ai-planning" ? AI_PLANNING_KEYWORDS : undefined),
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

  if (feature.slug === "ai-planning") {
    return <AiPlanningPage detail={feature} />;
  }

  const featurePageConfig = getFeaturePageConfig(feature.slug);

  if (featurePageConfig) {
    return <FeatureLandingPage config={featurePageConfig} detail={feature} />;
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
