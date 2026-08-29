import type { FAQPage, WebPage, WithContext } from "schema-dts";
import { Box, Text } from "ui";
import { CallToAction } from "@/components/shared/cta";
import { FeatureDetailHero } from "@/components/shared/feature-detail-hero";
import type { MarketingDetail } from "@/components/shared/marketing-detail-page";
import { Container } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import {
  FEATURE_STORY_META_TEXT_CLASS,
  FEATURE_STORY_SURFACE_CLASS,
  FEATURE_STORY_TEXT_CLASS,
} from "@/modules/home/feature-story-section";
import { ShowcaseCard } from "@/modules/home/decide-what-matters-showcase";
import { MayaWorkPlanCard } from "@/modules/home/how-it-works";
import mayaDeliveryBriefDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import { AiPlanningWorkflow } from "./ai-planning-workflow";

const PLANNING_CARD_TEXTURE = "/images/textures/decide-risograph.webp";

function DeliveryRiskCard() {
  return (
    <Box aria-hidden="true" className="flex h-full flex-col gap-3">
      <Box className={`${FEATURE_STORY_SURFACE_CLASS} px-4 py-3`}>
        <Box className="flex items-center justify-between gap-3">
          <Box>
            <Text
              className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}
            >
              Delivery brief
            </Text>
            <Text
              className={`${FEATURE_STORY_TEXT_CLASS} text-foreground font-semibold`}
            >
              Objectives needing attention
            </Text>
          </Box>
          <Text
            className={`${FEATURE_STORY_META_TEXT_CLASS} bg-danger/10 text-danger shrink-0 rounded-lg px-2.5 py-1 font-semibold`}
          >
            2 risks
          </Text>
        </Box>
      </Box>

      <Box className={`${FEATURE_STORY_SURFACE_CLASS} grid gap-2.5 p-4`}>
        <Box className="bg-surface-muted rounded-lg px-3 py-2.5">
          <Box className="flex items-center gap-2">
            <Box className="bg-danger size-2 shrink-0 rounded-full" />
            <Text
              className={`${FEATURE_STORY_TEXT_CLASS} text-foreground font-semibold`}
            >
              Improve product reliability
            </Text>
          </Box>
          <Text
            className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted mt-1 ml-4`}
          >
            API review is blocking delivery
          </Text>
        </Box>
        <Box className="bg-surface-muted rounded-lg px-3 py-2.5">
          <Box className="flex items-center gap-2">
            <Box className="bg-warning size-2 shrink-0 rounded-full" />
            <Text
              className={`${FEATURE_STORY_TEXT_CLASS} text-foreground font-semibold`}
            >
              Improve customer adoption
            </Text>
          </Box>
          <Text
            className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted mt-1 ml-4`}
          >
            Owner capacity needs a decision
          </Text>
        </Box>
      </Box>

      <Box className={`${FEATURE_STORY_SURFACE_CLASS} mt-auto px-4 py-3`}>
        <Text className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}>
          Protect reliability capacity before adding more work.
        </Text>
      </Box>
    </Box>
  );
}

function ReviewControlCard() {
  return (
    <Box aria-hidden="true" className="flex h-full flex-col gap-3">
      <Box className={`${FEATURE_STORY_SURFACE_CLASS} px-4 py-3`}>
        <Box className="flex items-center justify-between gap-3">
          <Box className="min-w-0">
            <Text
              className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}
            >
              Maya&apos;s recommendation
            </Text>
            <Text
              className={`${FEATURE_STORY_TEXT_CLASS} text-foreground truncate font-semibold`}
            >
              Plan onboarding research
            </Text>
          </Box>
          <Text
            className={`${FEATURE_STORY_META_TEXT_CLASS} bg-warning/10 text-warning shrink-0 rounded-lg px-2.5 py-1 font-semibold`}
          >
            Review
          </Text>
        </Box>
      </Box>

      <Box className={`${FEATURE_STORY_SURFACE_CLASS} grid gap-3 p-4`}>
        <Box className="border-border flex items-center justify-between border-b pb-3">
          <Text className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}>
            Suggested owner
          </Text>
          <Text
            className={`${FEATURE_STORY_TEXT_CLASS} text-foreground font-semibold`}
          >
            Product team
          </Text>
        </Box>
        <Box className="border-border flex items-center justify-between border-b pb-3">
          <Text className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}>
            Effort
          </Text>
          <Text
            className={`${FEATURE_STORY_TEXT_CLASS} text-foreground font-semibold`}
          >
            3 days
          </Text>
        </Box>
        <Box className="flex items-center justify-between gap-4">
          <Text className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}>
            Risk to resolve
          </Text>
          <Text
            className={`${FEATURE_STORY_META_TEXT_CLASS} text-foreground text-right font-semibold`}
          >
            Research access
          </Text>
        </Box>
      </Box>

      <Box className="mt-auto grid grid-cols-2 gap-2">
        <Text
          as="span"
          className={`${FEATURE_STORY_SURFACE_CLASS} ${FEATURE_STORY_META_TEXT_CLASS} text-foreground px-3 py-2.5 text-center font-semibold`}
        >
          Edit
        </Text>
        <Text
          as="span"
          className={`${FEATURE_STORY_META_TEXT_CLASS} bg-background-inverse text-foreground-inverse rounded-xl px-3 py-2.5 text-center font-semibold`}
        >
          Approve
        </Text>
      </Box>
    </Box>
  );
}

function PlanningDecisions() {
  return (
    <Container
      aria-labelledby="planning-decisions-title"
      as="section"
      className="scroll-mt-24 py-16 md:py-36"
    >
      <Box className="max-w-3xl" data-landing-reveal>
        <Text
          as="h2"
          className="text-3xl md:text-5xl"
          id="planning-decisions-title"
        >
          Ask what needs attention before the deadline does.
        </Text>
        <Text className="text-text-description mt-6 max-w-xl text-base text-pretty">
          Maya brings delivery risk, team capacity, and the proposed change into
          one reviewable planning decision.
        </Text>
      </Box>

      <Box className="mt-14 grid grid-cols-1 gap-x-6 gap-y-14 md:grid-cols-2 xl:grid-cols-3">
        <ShowcaseCard
          description="Bring at-risk objectives, blocked dependencies, and missing decisions forward while there is still time to act."
          illustrationClassName="max-w-[22rem]"
          imageSrc={PLANNING_CARD_TEXTURE}
          title="Surface delivery risk early."
        >
          <DeliveryRiskCard />
        </ShowcaseCard>
        <ShowcaseCard
          delay={70}
          description="Let Maya consider workload and availability before proposing an owner, effort, and first work window."
          illustrationClassName="max-w-[22rem]"
          imageSrc={PLANNING_CARD_TEXTURE}
          title="Plan around real capacity."
        >
          <MayaWorkPlanCard />
        </ShowcaseCard>
        <ShowcaseCard
          className="md:col-span-2 md:w-full md:max-w-[26rem] md:justify-self-center xl:col-span-1 xl:max-w-none"
          delay={140}
          description="Edit, approve, or reject important recommendations before they change ownership, timing, or scope."
          illustrationClassName="max-w-[22rem]"
          imageSrc={PLANNING_CARD_TEXTURE}
          title="Keep people in control."
        >
          <ReviewControlCard />
        </ShowcaseCard>
      </Box>
    </Container>
  );
}

export function AiPlanningPage({ detail }: { detail: MarketingDetail }) {
  const planningFaqs = detail.questions.map(([question, answer]) => ({
    answer,
    question,
  }));
  const pageJsonLd: WithContext<WebPage> = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    name: detail.heroTitle,
    description: detail.metaDescription,
    url: "https://www.fortyone.app/features/ai-planning",
    publisher: { "@type": "Organization", name: "FortyOne" },
  };
  const faqJsonLd: WithContext<FAQPage> = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: planningFaqs.map(({ question, answer }) => ({
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
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(pageJsonLd),
        }}
        type="application/ld+json"
      />
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }}
        type="application/ld+json"
      />
      <main className="bg-background text-foreground [&_h1]:font-semibold [&_h2]:font-semibold">
        <FeatureDetailHero
          description="Maya brings goals, capacity, timing, and risk into one clear proposal—before anyone commits the team."
          imageAlt="FortyOne Maya answering a delivery question with project momentum, completion trends, and the next planning prompt"
          imageDark={mayaDeliveryBriefDark}
          imageLight={mayaDeliveryBriefLight}
          title="Plan your team’s next move with Maya."
          url="https://fortyone.app/my-work"
        />
        <AiPlanningWorkflow />
        <PlanningDecisions />
        <Faqs
          heading="Frequently asked questions about planning with Maya"
          headingClassName="mx-auto max-w-2xl text-balance"
          items={planningFaqs}
          variant="pricing"
        />
      </main>
      <CallToAction
        className="border-t-0"
        contentClassName="pt-24 md:pt-32"
        description="Start free and ask Maya about the work already in your plan. No card and no trial clock."
        title="Make the next planning decision with context."
      />
    </>
  );
}
