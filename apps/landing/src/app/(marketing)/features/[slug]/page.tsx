import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { FeatureDetailHero } from "@/components/shared/feature-detail-hero";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { features, getFeatureBySlug } from "@/lib/features";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";
import mayaDeliveryBriefDark from "../../../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../../../public/images/product/maya-delivery-brief-light.webp";

const AI_PLANNING_HERO = (
  <FeatureDetailHero
    description="Maya brings goals, capacity, timing, and risk into one clear proposal—before anyone commits the team."
    imageAlt="FortyOne Maya answering a delivery question with project momentum, completion trends, and the next planning prompt"
    imageDark={mayaDeliveryBriefDark}
    imageLight={mayaDeliveryBriefLight}
    title="Plan your team’s next move with Maya."
    url="https://fortyone.app/my-work"
  />
);

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
      hero={feature.slug === "ai-planning" ? AI_PLANNING_HERO : undefined}
      questionHeading={`Questions about ${feature.label.toLowerCase()}`}
    />
  );
}
