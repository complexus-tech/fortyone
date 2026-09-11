import type { FAQPage, WebPage, WithContext } from "schema-dts";
import { Box, Text } from "ui";
import previewStyles from "@/modules/features/marketing-visual-card.module.css";
import { CallToAction } from "@/components/shared/cta";
import { FeatureDetailHero } from "@/components/shared/feature-detail-hero";
import type { MarketingDetail } from "@/components/shared/marketing-detail-page";
import { Container } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import { FeatureProductWorkflow } from "@/modules/features/feature-product-workflow";
import { CustomerStories } from "@/modules/home/customer-stories";
import { MarketingVisualCard } from "@/modules/features/marketing-visual-card";
import {
  ShowcaseCard,
  ShowcaseHeading,
} from "@/modules/home/decide-what-matters-showcase";
import { UseCaseHandoffs } from "./use-case-handoffs";
import { getHandoffConfig } from "./handoff-config";
import type { UseCasePageConfig } from "./use-case-page-config";

const USE_CASE_CARD_TEXTURE = "/images/textures/decide-risograph.webp";

function UseCaseOverview({ detail }: { detail: MarketingDetail }) {
  return (
    <Container
      aria-labelledby={`${detail.slug}-overview-title`}
      as="section"
      className="grid gap-10 py-16 md:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] md:gap-20 md:py-32"
    >
      <Box data-landing-reveal>
        <Text
          as="h2"
          className="max-w-lg text-3xl md:text-5xl"
          id={`${detail.slug}-overview-title`}
        >
          Built for your team’s everyday work.
        </Text>
      </Box>
      <Box className="max-w-2xl" data-landing-reveal>
        {detail.intro.map((paragraph) => (
          <Text
            className="text-text-description mb-6 text-base leading-relaxed text-pretty last:mb-0 md:text-lg"
            key={paragraph}
          >
            {paragraph}
          </Text>
        ))}
      </Box>
    </Container>
  );
}

function UseCaseDecisions({
  config,
  detail,
}: {
  config: UseCasePageConfig;
  detail: MarketingDetail;
}) {
  const sectionVisuals = detail.sections.flatMap(
    (section) => section.cards ?? [],
  );
  const visuals = [...detail.previewCards, ...sectionVisuals].slice(0, 3);
  const cards = detail.benefits
    .slice(0, 3)
    .map(([title, description], index) => ({
      description,
      title,
      visual: visuals[index],
    }));

  return (
    <Container
      aria-labelledby={`${detail.slug}-decisions-title`}
      as="section"
      className="py-16 md:py-32"
    >
      <ShowcaseHeading
        description={config.decisionDescription}
        id={`${detail.slug}-decisions-title`}
        title={config.decisionHeading}
      />

      <Box
        className={`${previewStyles.gallery} mt-14 grid grid-cols-1 gap-x-8 gap-y-16 md:grid-cols-2 xl:grid-cols-3 xl:gap-x-10`}
      >
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
              illustrationClassName="max-w-[20rem]"
              imageSrc={USE_CASE_CARD_TEXTURE}
              key={title}
              title={title}
              tone={(["sky", "sage", "amber"] as const)[index % 3]}
            >
              <MarketingVisualCard
                kind={index === 0 ? "feedback" : "planning"}
                visual={visual}
              />
            </ShowcaseCard>
          ) : null,
        )}
      </Box>
    </Container>
  );
}

export function UseCaseLandingPage({
  config,
  detail,
}: {
  config: UseCasePageConfig;
  detail: MarketingDetail;
}) {
  const faqs = detail.questions.map(([question, answer]) => ({
    answer,
    question,
  }));
  const workflowItems = config.workflow.visuals.map((visual, index) => {
    const section = detail.sections[index];

    return {
      ...visual,
      description:
        section?.paragraphs[0] ??
        "Keep the source, owner, and next decision connected as work moves.",
      title: section?.title ?? visual.label,
    };
  });
  const pageJsonLd: WithContext<WebPage> = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    name: config.hero.title,
    description: detail.metaDescription,
    url: `https://www.fortyone.app/use-cases/${detail.slug}`,
    publisher: { "@type": "Organization", name: "FortyOne" },
  };
  const faqJsonLd: WithContext<FAQPage> = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: faqs.map(({ question, answer }) => ({
      "@type": "Question",
      name: question,
      acceptedAnswer: { "@type": "Answer", text: answer },
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
      <main className="bg-background text-foreground [&_h1]:font-semibold [&_h2]:font-semibold [&_h3]:font-semibold">
        <FeatureDetailHero
          description={config.hero.description}
          imageAlt={config.hero.alt}
          imageDark={config.hero.darkImage}
          imageLight={config.hero.lightImage}
          title={config.hero.title}
          url={config.hero.url}
        />
        <UseCaseOverview detail={detail} />
        <FeatureProductWorkflow
          ariaLabel={`Explore the ${detail.label.toLowerCase()} workflow`}
          description={config.workflow.description}
          heading={config.workflow.heading}
          id={`${detail.slug}-workflow`}
          items={workflowItems}
        />
        <UseCaseDecisions config={config} detail={detail} />
        <UseCaseHandoffs
          config={getHandoffConfig(detail.slug)}
          detail={{
            slug: detail.slug,
            label: detail.label,
            sections: detail.sections,
          }}
          key={detail.slug}
        />
        <CustomerStories />
        <Faqs
          heading={`Frequently asked questions from ${detail.label.toLowerCase()} teams`}
          headingClassName="mx-auto max-w-2xl text-balance"
          items={faqs}
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
