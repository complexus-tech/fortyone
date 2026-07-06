import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { features, getFeatureBySlug } from "@/lib/features";

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

  const canonicalUrl = `https://www.fortyone.app/features/${feature.slug}`;

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
    },
    twitter: {
      card: "summary_large_image",
      title: feature.metaTitle,
      description: feature.metaDescription,
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
