import type { FAQPage, WebPage, WithContext } from "schema-dts";
import { Box, Text } from "ui";
import { CallToAction } from "@/components/shared/cta";
import { FeatureDetailHero } from "@/components/shared/feature-detail-hero";
import type { MarketingDetail } from "@/components/shared/marketing-detail-page";
import { Container } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import { ShowcaseCard } from "@/modules/home/decide-what-matters-showcase";
import type { FeatureLandingConfig } from "./feature-page-config";
import { FeatureProductWorkflow } from "./feature-product-workflow";
import { MarketingVisualCard } from "./marketing-visual-card";

const FEATURE_CARD_TEXTURE = "/images/textures/decide-risograph.webp";

function FeatureDecisions({
  config,
  detail,
}: {
  config: FeatureLandingConfig;
  detail: MarketingDetail;
}) {
  const sectionVisuals = detail.sections.flatMap(
    (section) => section.cards ?? [],
  );
  const visuals = [...detail.previewCards, ...sectionVisuals].slice(0, 3);
  const cards = config.decisionCards.map(({ description, title }, index) => ({
    description,
    title,
    visual: visuals[index],
  }));

  return (
    <Container
      aria-labelledby={`${detail.slug}-decisions-title`}
      as="section"
      className="scroll-mt-24 py-16 md:py-36"
    >
      <Box className="max-w-3xl" data-landing-reveal>
        <Text
          as="h2"
          className="text-3xl md:text-5xl"
          id={`${detail.slug}-decisions-title`}
        >
          {config.decisionHeading}
        </Text>
        <Text className="text-text-description mt-6 max-w-xl text-base text-pretty">
          {config.decisionDescription}
        </Text>
      </Box>

      <Box className="mt-14 grid grid-cols-1 gap-x-6 gap-y-14 md:grid-cols-2 xl:grid-cols-3">
        {cards.map(({ description, title, visual }, index) =>
          visual ? (
            <ShowcaseCard
              className={
                index === 2
                  ? "md:col-span-2 md:w-full md:max-w-[26rem] md:justify-self-center xl:col-span-1 xl:max-w-none"
                  : undefined
              }
              delay={index * 70}
              description={description}
              illustrationClassName="max-w-[22rem]"
              imageSrc={FEATURE_CARD_TEXTURE}
              key={title}
              title={title}
            >
              <MarketingVisualCard visual={visual} />
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
        <FeatureProductWorkflow {...config.workflow} />
        <FeatureDecisions config={config} detail={detail} />
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
