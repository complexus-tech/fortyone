import type { FAQPage, WebPage, WithContext } from "schema-dts";
import { Box, Text } from "ui";
import { CallToAction } from "@/components/shared/cta";
import { FeatureDetailHero } from "@/components/shared/feature-detail-hero";
import type {
  MarketingDetail,
  MarketingSection,
} from "@/components/shared/marketing-detail-page";
import { Container } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import { FeatureProductWorkflow } from "@/modules/features/feature-product-workflow";
import { MarketingVisualCard } from "@/modules/features/marketing-visual-card";
import { ShowcaseCard } from "@/modules/home/decide-what-matters-showcase";
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
          Built around the work {detail.label.toLowerCase()} teams already do.
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
              imageSrc={USE_CASE_CARD_TEXTURE}
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

function SectionRows({ rows }: { rows: [string, string][] }) {
  return (
    <dl className="border-border mt-8 grid border-y sm:grid-cols-2">
      {rows.map(([label, value], index) => (
        <Box
          className={`py-5 sm:px-5 ${
            index % 2 === 0 ? "sm:border-border sm:border-r sm:pl-0" : "sm:pr-0"
          } ${index > 1 ? "border-border border-t" : ""}`}
          key={label}
        >
          <Text as="dt" className="font-semibold">
            {label}
          </Text>
          <Text as="dd" className="text-text-muted mt-2 text-sm leading-6">
            {value}
          </Text>
        </Box>
      ))}
    </dl>
  );
}

function UseCaseComparison({
  comparison,
}: {
  comparison: NonNullable<MarketingSection["comparisonTable"]>;
}) {
  return (
    <Box className="border-border mt-8 overflow-x-auto border-y">
      <table className="w-full min-w-[620px] text-left text-sm">
        <thead>
          <tr className="border-border border-b">
            <th className="py-4 pr-5 font-semibold">Capability</th>
            <th className="py-4 pr-5 text-center font-semibold">FortyOne</th>
            <th className="py-4 pr-5 text-center font-semibold">
              {comparison.competitor}
            </th>
            <th className="py-4 font-semibold">Planning impact</th>
          </tr>
        </thead>
        <tbody>
          {comparison.rows.map((row) => (
            <tr
              className="border-border border-b last:border-b-0"
              key={row.feature}
            >
              <td className="py-4 pr-5 font-medium">{row.feature}</td>
              <td className="py-4 pr-5 text-center">
                {row.fortyOne ? "Included" : "—"}
              </td>
              <td className="text-text-muted py-4 pr-5 text-center">
                {row.competitor ? "Included" : "—"}
              </td>
              <td className="text-text-muted py-4 leading-6">{row.note}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Box>
  );
}

function UseCaseDetails({ detail }: { detail: MarketingDetail }) {
  return (
    <Container
      aria-labelledby={`${detail.slug}-details-title`}
      as="section"
      className="py-16 md:py-32"
    >
      <Box className="max-w-3xl" data-landing-reveal>
        <Text
          as="h2"
          className="text-3xl md:text-5xl"
          id={`${detail.slug}-details-title`}
        >
          Keep the context behind every handoff.
        </Text>
        <Text className="text-text-description mt-6 max-w-xl text-base text-pretty">
          The plan stays useful because the request, decision, owner, and
          delivery evidence remain connected as the work moves.
        </Text>
      </Box>

      <Box className="mt-14 border-t">
        {detail.sections.map((section, index) => (
          <Box
            className="border-border grid gap-8 border-b py-10 md:grid-cols-[minmax(0,0.72fr)_minmax(0,1.28fr)] md:gap-16 md:py-14"
            data-landing-reveal
            id={section.id}
            key={section.id}
          >
            <Box>
              <Text className="text-text-muted text-sm tabular-nums">
                0{index + 1}
              </Text>
              <Text as="h3" className="mt-3 max-w-md text-2xl md:text-3xl">
                {section.title}
              </Text>
            </Box>
            <Box className="max-w-2xl">
              {section.paragraphs.map((paragraph) => (
                <Text
                  className="text-text-description mb-5 text-base leading-relaxed last:mb-0"
                  key={paragraph}
                >
                  {paragraph}
                </Text>
              ))}
              {section.rows ? <SectionRows rows={section.rows} /> : null}
              {section.comparisonTable ? (
                <UseCaseComparison comparison={section.comparisonTable} />
              ) : null}
              {section.prompt && section.promptTitle ? (
                <Box className="bg-surface-muted mt-8 rounded-2xl p-5 md:p-6">
                  <Text className="text-text-muted text-sm font-semibold">
                    {section.promptTitle}
                  </Text>
                  <Text className="mt-3 leading-7">{section.prompt}</Text>
                </Box>
              ) : null}
            </Box>
          </Box>
        ))}
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
        <UseCaseDetails detail={detail} />
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
