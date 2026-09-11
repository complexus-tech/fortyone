import type { FAQPage, WebPage, WithContext } from "schema-dts";
import { Box } from "ui";
import { CallToAction } from "@/components/shared/cta";
import { FeatureDetailHero } from "@/components/shared/feature-detail-hero";
import type { MarketingDetail } from "@/components/shared/marketing-detail-page";
import { Container } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import {
  ShowcaseCard,
  ShowcaseHeading,
} from "@/modules/home/decide-what-matters-showcase";
import { IntegrationDirectory } from "@/modules/integrations/integration-directory";
import { CustomerStories } from "@/modules/home/customer-stories";
import previewStyles from "./marketing-visual-card.module.css";
import type { FeatureLandingConfig } from "./feature-page-config";
import { FeatureProductWorkflow } from "./feature-product-workflow";
import { getFeatureIllustrations } from "./feature-illustrations";
import { MarketingVisualCard } from "./marketing-visual-card";

const FEATURE_CARD_TEXTURE = "/images/textures/decide-risograph.webp";

function FeatureDecisions({
  config,
  detail,
}: {
  config: FeatureLandingConfig;
  detail: MarketingDetail;
}) {
  const illustrations = getFeatureIllustrations(detail.slug);
  const cards = config.decisionCards.map(({ description, title }, index) => ({
    description,
    title,
    illustration: illustrations?.[index],
  }));

  return (
    <Container
      aria-labelledby={`${detail.slug}-decisions-title`}
      as="section"
      className="scroll-mt-24 py-16 md:py-36"
    >
      <ShowcaseHeading
        description={config.decisionDescription}
        id={`${detail.slug}-decisions-title`}
        title={config.decisionHeading}
        titleClassName="max-w-xl text-balance"
      />

      <Box
        className={`${previewStyles.gallery} mt-14 grid grid-cols-1 gap-x-8 gap-y-16 md:grid-cols-2 xl:grid-cols-3 xl:gap-x-10`}
      >
        {cards.map(({ description, title, illustration }, index) =>
          illustration ? (
            <ShowcaseCard
              className={
                index === 2
                  ? "md:col-span-2 md:w-full md:max-w-[26rem] md:justify-self-center xl:col-span-1 xl:max-w-none"
                  : undefined
              }
              delay={index * 70}
              description={description}
              illustrationClassName="max-w-[20rem]"
              imageSrc={FEATURE_CARD_TEXTURE}
              key={title}
              title={title}
              tone={(["sky", "sage", "amber"] as const)[index % 3]}
            >
              <MarketingVisualCard {...illustration} />
            </ShowcaseCard>
          ) : null,
        )}
      </Box>
    </Container>
  );
}

export function FeatureLandingPage({
  config,
  detail,
}: {
  config: FeatureLandingConfig;
  detail: MarketingDetail;
}) {
  const featureFaqs = detail.questions.map(([question, answer]) => ({
    answer,
    question,
  }));
  const pageJsonLd: WithContext<WebPage> = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    name: detail.heroTitle,
    description: detail.metaDescription,
    url: `https://www.fortyone.app/features/${detail.slug}`,
    publisher: { "@type": "Organization", name: "FortyOne" },
  };
  const faqJsonLd: WithContext<FAQPage> = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: featureFaqs.map(({ question, answer }) => ({
      "@type": "Question",
      name: question,
      acceptedAnswer: {
        "@type": "Answer",
        text: answer,
      },
    })),
  };

  return (
    <>
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(pageJsonLd) }}
        type="application/ld+json"
      />
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }}
        type="application/ld+json"
      />
      <main className="bg-background text-foreground [&_h1]:font-semibold [&_h2]:font-semibold">
        <FeatureDetailHero
          description={config.hero.description}
          imageAlt={config.hero.alt}
          imageDark={config.hero.darkImage}
          imageLight={config.hero.lightImage}
          title={config.hero.title}
          url={config.hero.url}
        />
        {detail.slug === "integrations" ? <IntegrationDirectory /> : null}
        <FeatureProductWorkflow {...config.workflow} />
        <FeatureDecisions config={config} detail={detail} />
        <CustomerStories />
        <Faqs
          heading={config.faqHeading}
          headingClassName="mx-auto max-w-2xl text-balance"
          items={featureFaqs}
          variant="pricing"
        />
      </main>
      <CallToAction
        className="border-t-0"
        contentClassName="pt-24 md:pt-32"
        description={config.ctaDescription}
        title={config.ctaTitle}
      />
    </>
  );
}
