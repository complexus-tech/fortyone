import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { getUseCaseBySlug, useCases } from "@/lib/use-cases";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";

export function generateStaticParams() {
  return useCases.map((useCase) => ({ slug: useCase.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const useCase = getUseCaseBySlug(slug);

  if (!useCase) {
    return {};
  }

  const canonicalUrl = getCanonicalUrl(`/use-cases/${useCase.slug}`);

  return {
    title: useCase.metaTitle,
    description: useCase.metaDescription,
    alternates: {
      canonical: canonicalUrl,
    },
    openGraph: {
      type: "website",
      url: canonicalUrl,
      title: useCase.metaTitle,
      description: useCase.metaDescription,
      siteName: "FortyOne",
      images: [DEFAULT_SOCIAL_IMAGE],
    },
    twitter: {
      card: "summary_large_image",
      title: useCase.metaTitle,
      description: useCase.metaDescription,
      images: [DEFAULT_TWITTER_IMAGE],
    },
  };
}

export default async function UseCasePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const useCase = getUseCaseBySlug(slug);

  if (!useCase) {
    return notFound();
  }

  return (
    <MarketingDetailPage
      basePath="use-cases"
      breadcrumbLabel="Use cases"
      detail={useCase}
      questionHeading={`Questions from ${useCase.label.toLowerCase()}`}
    />
  );
}
