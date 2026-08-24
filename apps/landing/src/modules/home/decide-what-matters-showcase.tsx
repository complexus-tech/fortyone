import type { ReactNode } from "react";
import Image from "next/image";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import {
  IntegrationCard,
  MayaWorkPlanCard,
  RequestToWorkCard,
} from "./how-it-works";

type ShowcaseCardProps = {
  children: ReactNode;
  className?: string;
  delay?: number;
  description: string;
  imageSrc: string;
  illustrationClassName?: string;
  title: string;
};

const ShowcaseCard = ({
  children,
  className,
  delay = 0,
  description,
  imageSrc,
  illustrationClassName,
  title,
}: ShowcaseCardProps) => {
  return (
    <Box
      as="article"
      className={cn("min-w-0", className)}
      data-landing-reveal
      style={{ transitionDelay: `${delay}ms` }}
    >
      <Box className="relative aspect-[5/6] overflow-hidden rounded-xl">
        <Image
          alt=""
          className="object-cover object-bottom"
          fill
          sizes="(max-width: 767px) 100vw, (max-width: 1279px) 50vw, 33vw"
          src={imageSrc}
        />
        <Box className="relative z-10 flex h-full items-center justify-center p-5 sm:p-8 md:p-6 xl:p-8">
          <Box
            aria-hidden="true"
            className={cn(
              "[&_.bg-surface-elevated]:bg-surface-elevated/95 dark:[&_.bg-surface-elevated]:bg-surface-prominent/90 w-full [&_.bg-surface-elevated]:border-transparent [&_.bg-surface-elevated]:backdrop-blur-md",
              illustrationClassName,
            )}
          >
            {children}
          </Box>
        </Box>
      </Box>

      <Box className="mt-6">
        <Text as="h3" className="text-foreground mb-2 text-lg font-semibold">
          {title}
        </Text>
        <Text className="text-text-muted leading-relaxed">{description}</Text>
      </Box>
    </Box>
  );
};

export const DecideWhatMattersShowcase = () => {
  return (
    <Container
      aria-labelledby="ai-planning-title"
      as="section"
      className="scroll-mt-24 pt-16 pb-8 md:pt-36"
      id="ai-planning"
    >
      <Box>
        <Box className="max-w-4xl" data-landing-reveal>
          <Text as="h2" className="text-3xl md:text-5xl" id="ai-planning-title">
            Choose the right work. Build a plan your team can deliver.
          </Text>
          <Text className="text-text-muted mt-6 max-w-2xl text-pretty">
            Bring goals, customer feedback, capacity, and delivery context into
            one place, then let Maya turn the right priorities into a realistic
            plan.
          </Text>
        </Box>

        <Box className="mt-14 grid grid-cols-1 gap-x-6 gap-y-14 md:grid-cols-2 xl:grid-cols-3">
          <ShowcaseCard
            description="Connect goals to customer feedback so the work with the strongest case rises to the top."
            illustrationClassName="max-w-[22rem]"
            imageSrc="/images/textures/decide-risograph.png"
            title="Prioritise with evidence."
          >
            <RequestToWorkCard density="comfortable" />
          </ShowcaseCard>

          <ShowcaseCard
            delay={70}
            description="Maya checks workload and calendars to suggest an owner and delivery window the team can actually commit to."
            illustrationClassName="max-w-[22rem]"
            imageSrc="/images/textures/decide-risograph.png"
            title="Plan around real capacity."
          >
            <MayaWorkPlanCard density="comfortable" />
          </ShowcaseCard>

          <ShowcaseCard
            className="md:col-span-2 md:w-full md:max-w-[26rem] md:justify-self-center xl:col-span-1 xl:max-w-none"
            delay={140}
            description="Goals, requests, documents, conversations, and delivery links stay attached from decision to done."
            illustrationClassName="max-w-[22rem]"
            imageSrc="/images/textures/decide-risograph.png"
            title="Keep context with the work."
          >
            <IntegrationCard density="comfortable" />
          </ShowcaseCard>
        </Box>
      </Box>
    </Container>
  );
};
