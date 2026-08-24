import { Box, Text } from "ui";
import { CallToAction } from "@/components/shared/cta";
import { FeatureDetailHero } from "@/components/shared/feature-detail-hero";
import type { MarketingDetail } from "@/components/shared/marketing-detail-page";
import { Container } from "@/components/ui";
import { ProductScreenshot } from "@/modules/home/product-screenshot";
import mayaDeliveryBriefDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import mayaHomeDark from "../../../public/images/product/maya-home-dark.webp";
import mayaHomeLight from "../../../public/images/product/maya-home-light.webp";
import mayaObjectiveRisksDark from "../../../public/images/product/maya-objective-risks-dark.webp";
import mayaObjectiveRisksLight from "../../../public/images/product/maya-objective-risks-light.webp";

const PLANNING_STEPS = [
  [
    "Bring the right context",
    "Ask Maya from the work already in FortyOne, with goals, tasks, workload, and connected project context close at hand.",
  ],
  [
    "Prepare a grounded next move",
    "Turn that context into a proposed owner, effort, start window, and the risks the team should consider.",
  ],
  [
    "Review before anything changes",
    "Keep managers in control of the decision. Adjust, approve, or reject the recommendation before it touches the plan.",
  ],
] as const;

function PlanningOverview() {
  return (
    <section
      aria-labelledby="planning-overview-title"
      className="pt-16 md:pt-36"
    >
      <Container>
        <Box className="max-w-4xl" data-landing-reveal>
          <Text
            as="h2"
            className="text-3xl md:text-5xl"
            id="planning-overview-title"
          >
            Give Maya the context. Get a plan your team can stand behind.
          </Text>
          <Text className="text-text-muted mt-6 max-w-2xl text-pretty">
            Start with a question in plain language. Maya brings the work around
            it into view, then helps the team move from context to a decision.
          </Text>
        </Box>
      </Container>

      <ProductScreenshot
        alt="Maya AI project agent ready to answer questions about the team, assigned work, notifications, and priorities"
        containerClassName="mt-10 md:mt-16"
        darkImage={mayaHomeDark}
        lightImage={mayaHomeLight}
        url="https://fortyone.app/maya"
      />

      <Container>
        <Box className="mt-12 grid gap-10 md:grid-cols-3 md:gap-8">
          {PLANNING_STEPS.map(([title, description], index) => (
            <Box
              className="border-border border-t pt-5"
              data-landing-reveal
              key={title}
              style={{ transitionDelay: `${index * 70}ms` }}
            >
              <Text className="text-primary mb-7 text-sm font-semibold">
                0{index + 1}
              </Text>
              <Text as="h3" className="text-foreground text-lg font-semibold">
                {title}
              </Text>
              <Text className="text-text-muted mt-3 leading-relaxed">
                {description}
              </Text>
            </Box>
          ))}
        </Box>
      </Container>
    </section>
  );
}

function RiskStory() {
  return (
    <section aria-labelledby="planning-risk-title" className="py-16 md:py-36">
      <Container>
        <Box className="max-w-4xl" data-landing-reveal>
          <Text className="text-primary mb-5 text-sm font-semibold tracking-wide uppercase">
            Protect delivery
          </Text>
          <Text
            as="h2"
            className="text-3xl md:text-5xl"
            id="planning-risk-title"
          >
            Ask what needs attention before the deadline does.
          </Text>
          <Text className="text-text-muted mt-6 max-w-2xl leading-relaxed text-pretty">
            Maya can bring live work and goal health into the conversation,
            highlight the objectives that need a decision, and explain the
            tradeoff before the team commits to more work.
          </Text>
        </Box>
      </Container>

      <ProductScreenshot
        alt="FortyOne summary with Maya highlighting two at-risk objectives and the delivery decision the team should make"
        containerClassName="mt-10 md:mt-16"
        darkImage={mayaObjectiveRisksDark}
        lightImage={mayaObjectiveRisksLight}
        url="https://fortyone.app/summary"
      />
    </section>
  );
}

function Benefits({ benefits }: { benefits: MarketingDetail["benefits"] }) {
  return (
    <Container as="section" className="py-16 md:py-28">
      <Box className="landing-hero-shell overflow-hidden rounded-[2.5rem] px-6 py-12 sm:px-10 md:rounded-[3.5rem] md:px-14 md:py-18 xl:px-20">
        <Box className="max-w-3xl" data-landing-reveal>
          <Text as="h2" className="text-3xl md:text-5xl">
            Move faster without losing control of the plan.
          </Text>
        </Box>
        <Box className="border-border bg-border mt-12 grid gap-px overflow-hidden rounded-2xl border sm:grid-cols-2 xl:grid-cols-4">
          {benefits.map(([title, description], index) => (
            <Box
              className="bg-background min-h-52 p-6 sm:p-7"
              data-landing-reveal
              key={title}
              style={{ transitionDelay: `${index * 60}ms` }}
            >
              <Text className="text-primary mb-8 text-sm font-semibold">
                0{index + 1}
              </Text>
              <Text as="h3" className="text-foreground text-lg font-semibold">
                {title}
              </Text>
              <Text className="text-text-muted mt-3 text-sm leading-relaxed">
                {description}
              </Text>
            </Box>
          ))}
        </Box>
      </Box>
    </Container>
  );
}

function Questions({ questions }: { questions: MarketingDetail["questions"] }) {
  return (
    <Container as="section" className="py-16 md:py-28">
      <Box className="grid gap-10 md:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] md:gap-16 xl:gap-24">
        <Text
          as="h2"
          className="max-w-md text-3xl md:text-5xl"
          data-landing-reveal
        >
          Questions about planning with Maya.
        </Text>
        <Box className="border-border border-t">
          {questions.map(([question, answer], index) => (
            <Box
              className="border-border border-b py-7"
              data-landing-reveal
              key={question}
              style={{ transitionDelay: `${index * 60}ms` }}
            >
              <Text as="h3" className="text-foreground text-lg font-semibold">
                {question}
              </Text>
              <Text className="text-text-muted mt-3 max-w-2xl leading-relaxed">
                {answer}
              </Text>
            </Box>
          ))}
        </Box>
      </Box>
    </Container>
  );
}

export function AiPlanningPage({ detail }: { detail: MarketingDetail }) {
  return (
    <>
      <script
        dangerouslySetInnerHTML={{
          __html: JSON.stringify({
            "@context": "https://schema.org",
            "@type": "WebPage",
            name: detail.heroTitle,
            description: detail.metaDescription,
            url: "https://www.fortyone.app/features/ai-planning",
            publisher: { "@type": "Organization", name: "FortyOne" },
          }),
        }}
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
        <PlanningOverview />
        <RiskStory />
        <Benefits benefits={detail.benefits} />
        <Questions questions={detail.questions} />
      </main>
      <CallToAction />
    </>
  );
}
