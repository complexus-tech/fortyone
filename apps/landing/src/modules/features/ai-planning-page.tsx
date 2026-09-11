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
import { PlanningPreview } from "@/modules/home/capability-previews";
import { CustomerStories } from "@/modules/home/customer-stories";
import mayaDeliveryBriefDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import previewStyles from "./marketing-visual-card.module.css";
import { MarketingVisualCard } from "./marketing-visual-card";
import { AiPlanningWorkflow } from "./ai-planning-workflow";

const PLANNING_CARD_TEXTURE = "/images/textures/decide-risograph.webp";

function DeliveryRiskCard() {
  return (
    <MarketingVisualCard
      kind="planning"
      visual={{
        heading: "Delivery brief",
        subheading: "Objectives needing attention",
        badge: "2 risks",
        rows: [
          {
            label: "Product reliability",
            value: "API review is blocking delivery",
          },
          {
            label: "Customer adoption",
            value: "Owner capacity needs a decision",
          },
        ],
        note: "Resolve the risk before adding more work",
      }}
    />
  );
}

function ReviewControlCard() {
  return (
    <MarketingVisualCard
      kind="planning"
      visual={{
        heading: "Maya’s recommendation",
        subheading: "Plan onboarding research",
        badge: "Review",
        rows: [
          { label: "Suggested owner", value: "Product team" },
          {
            label: "Effort · Risk to resolve",
            value: "3 days · Research access",
          },
        ],
        note: "Edit, approve, or reject before applying",
      }}
    />
  );
}

function PlanningDecisions() {
  return (
    <Container
      aria-labelledby="planning-decisions-title"
      as="section"
      className="scroll-mt-24 py-16 md:py-36"
    >
      <ShowcaseHeading
        description="Maya brings delivery risk, team capacity, and the proposed change into one reviewable planning decision."
        id="planning-decisions-title"
        title="Ask what needs attention before the deadline does."
      />

      <Box
        className={`${previewStyles.gallery} mt-14 grid grid-cols-1 gap-x-8 gap-y-16 md:grid-cols-2 xl:grid-cols-3 xl:gap-x-10`}
      >
        <ShowcaseCard
          description="Bring at-risk objectives, blocked dependencies, and missing decisions forward while there is still time to act."
          illustrationClassName="max-w-[20rem]"
          imageSrc={PLANNING_CARD_TEXTURE}
          title="Surface delivery risk early."
          tone="sky"
        >
          <DeliveryRiskCard />
        </ShowcaseCard>
        <ShowcaseCard
          delay={70}
          description="Let Maya consider workload and availability before proposing an owner, effort, and first work window."
          illustrationClassName="max-w-[20rem]"
          imageSrc={PLANNING_CARD_TEXTURE}
          title="Plan around real capacity."
          tone="sage"
        >
          <PlanningPreview />
        </ShowcaseCard>
        <ShowcaseCard
          className="md:col-span-2 md:w-full md:max-w-[26rem] md:justify-self-center xl:col-span-1 xl:max-w-none"
          delay={140}
          description="Edit, approve, or reject important recommendations before they change ownership, timing, or scope."
          illustrationClassName="max-w-[20rem]"
          imageSrc={PLANNING_CARD_TEXTURE}
          title="Keep people in control."
          tone="amber"
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
        <CustomerStories />
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
