import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import {
  comparisons,
  getComparisonBySlug,
  getComparisonMarketingDetail,
} from "@/lib/comparisons";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";

export function generateStaticParams() {
  return comparisons.map((comparison) => ({ slug: comparison.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const comparison = getComparisonBySlug(slug);

  if (!comparison) {
    return {};
  }

  const canonicalUrl = getCanonicalUrl(`/compare/${comparison.slug}`);

  return {
    title: comparison.metaTitle,
    description: comparison.metaDescription,
    alternates: {
      canonical: canonicalUrl,
    },
    openGraph: {
      type: "website",
      url: canonicalUrl,
      title: comparison.metaTitle,
      description: comparison.metaDescription,
      siteName: "FortyOne",
      images: [DEFAULT_SOCIAL_IMAGE],
    },
    twitter: {
      card: "summary_large_image",
      title: comparison.metaTitle,
      description: comparison.metaDescription,
      images: [DEFAULT_TWITTER_IMAGE],
    },
  };
}

export default async function CompareDetailPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const comparison = getComparisonBySlug(slug);

  if (!comparison) {
    return notFound();
  }

  const detail = getComparisonMarketingDetail(comparison);

  return (
    <MarketingDetailPage
      basePath="compare"
      breadcrumbLabel="Compare"
      detail={detail}
      questionHeading={`Questions about FortyOne vs ${comparison.competitor}`}
    />
  );
}
