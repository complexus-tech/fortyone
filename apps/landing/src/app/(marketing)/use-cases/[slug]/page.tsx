import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MarketingDetailPage } from "@/components/shared/marketing-detail-page";
import { getUseCaseBySlug, useCases } from "@/lib/use-cases";

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

  const canonicalUrl = `https://www.fortyone.app/use-cases/${useCase.slug}`;

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
    },
    twitter: {
      card: "summary_large_image",
      title: useCase.metaTitle,
      description: useCase.metaDescription,
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
