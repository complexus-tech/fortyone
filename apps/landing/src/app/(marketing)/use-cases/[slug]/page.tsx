import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getUseCaseBySlug, useCases } from "@/lib/use-cases";
import { UseCaseLandingPage } from "@/modules/use-cases/use-case-landing-page";
import { getUseCasePageConfig } from "@/modules/use-cases/use-case-page-config";
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
  const pageConfig = getUseCasePageConfig(useCase.slug);

  return {
    title: useCase.metaTitle,
    description: useCase.metaDescription,
    keywords: pageConfig?.keywords,
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
  const pageConfig = getUseCasePageConfig(slug);

  if (!useCase || !pageConfig) {
    return notFound();
  }

  return <UseCaseLandingPage config={pageConfig} detail={useCase} />;
}
