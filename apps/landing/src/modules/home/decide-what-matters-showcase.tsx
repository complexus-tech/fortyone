import type { ReactNode } from "react";
import Image from "next/image";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import styles from "./capability-cards.module.css";
import {
  ContextPreview,
  FeedbackPreview,
  PlanningPreview,
} from "./capability-previews";

type ShowcaseCardProps = {
  children: ReactNode;
  className?: string;
  delay?: number;
  description: string;
  imageSrc: string;
  illustrationClassName?: string;
  title: string;
  tone?: "sky" | "sage" | "amber";
};

export const ShowcaseCard = ({
  children,
  className,
  delay = 0,
  description,
  imageSrc,
  illustrationClassName,
  title,
  tone,
}: ShowcaseCardProps) => {
  return (
    <Box
      as="article"
      className={cn("min-w-0", className)}
      data-landing-reveal
      style={{ transitionDelay: `${delay}ms` }}
    >
      <Box
        className={cn(
          "relative aspect-square overflow-hidden rounded-lg md:rounded-xl",
          tone && styles[tone],
        )}
      >
        <Image
          alt=""
          className={cn(
            "object-cover object-center",
            tone && "opacity-20 mix-blend-luminosity",
          )}
          fill
          sizes="(max-width: 767px) 100vw, (max-width: 1279px) 50vw, 33vw"
          src={imageSrc}
        />
        {tone ? (
          <svg
            aria-hidden="true"
            className={styles.illustration}
            fill="none"
            viewBox="0 0 400 400"
          >
            {tone === "sky" && (
              <>
                <path d="M-20 305C70 305 90 70 215 70S335 205 425 150" />
                <path d="M-20 335C80 335 110 100 225 100S345 235 425 180" />
                <circle cx="91" cy="215" r="8" />
                <circle cx="276" cy="89" r="6" />
                <path d="M325 285v22m-11-11h22M65 65v14m-7-7h14" />
              </>
            )}
            {tone === "sage" && (
              <>
                <circle cx="200" cy="205" r="116" />
                <circle cx="200" cy="205" r="157" />
                <circle cx="200" cy="205" r="195" />
                <circle cx="95" cy="88" r="7" />
                <circle cx="316" cy="310" r="5" />
                <path d="M310 65v22m-11-11h22" />
              </>
            )}
            {tone === "amber" && (
              <>
                <path d="M40 50h95l65 80h130v145l-90 70H65V205L40 50Z" />
                <path d="m40 50 160 80 40 215M65 205l135-75 130 145" />
                <circle cx="40" cy="50" r="7" />
                <circle cx="330" cy="275" r="7" />
                <circle cx="240" cy="345" r="5" />
                <path d="M307 52v20m-10-10h20" />
              </>
            )}
          </svg>
        ) : null}
        <Box
          className={cn(
            "relative z-10 flex h-full w-full items-center justify-center p-3 sm:p-8 md:p-6 xl:p-8",
            tone && "px-2 sm:px-4 md:px-3 xl:px-4",
          )}
        >
          <Box
            aria-hidden="true"
            className={cn(
              "w-full",
              !tone &&
                "[&_.bg-surface-elevated]:bg-surface-elevated/95 dark:[&_.bg-surface-elevated]:bg-surface-prominent/90 [&_.bg-surface-elevated]:border-transparent [&_.bg-surface-elevated]:backdrop-blur-md",
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
        <Text className="text-text-description text-base leading-relaxed">
          {description}
        </Text>
      </Box>
    </Box>
  );
};

export const DecideWhatMattersShowcase = () => {
  return (
    <Container
      aria-labelledby="ai-planning-title"
      as="section"
      className="scroll-mt-24 pt-20 pb-8 md:pt-32"
      id="ai-planning"
    >
      <Box>
        <Box
          className="grid gap-6 md:grid-cols-[1.2fr_1fr] md:items-end md:gap-24"
          data-landing-reveal
        >
          <Box>
            <Text
              as="h2"
              className="text-3xl md:text-5xl"
              id="ai-planning-title"
            >
              Decide with evidence. Plan around real capacity.
            </Text>
          </Box>
          <Text className="text-text-description max-w-lg text-base leading-relaxed text-pretty">
            Connect goals to customer demand, then let Maya recommend ownership
            and a delivery window before the team commits.
          </Text>
        </Box>

        <Box className="mt-14 grid grid-cols-1 gap-x-8 gap-y-16 md:grid-cols-2 xl:grid-cols-3 xl:gap-x-10">
          <ShowcaseCard
            description="Connect goals to customer feedback so the work with the strongest case rises to the top."
            illustrationClassName="max-w-[20rem]"
            imageSrc="/images/textures/decide-risograph.webp"
            title="Prioritize with evidence."
            tone="sky"
          >
            <FeedbackPreview />
          </ShowcaseCard>

          <ShowcaseCard
            delay={70}
            description="Maya checks workload and calendars to suggest an owner and delivery window the team can actually commit to."
            illustrationClassName="max-w-[20rem]"
            imageSrc="/images/textures/decide-risograph.webp"
            title="Plan around real capacity."
            tone="sage"
          >
            <PlanningPreview />
          </ShowcaseCard>

          <ShowcaseCard
            className="md:col-span-2 md:w-full md:max-w-[26rem] md:justify-self-center xl:col-span-1 xl:max-w-none"
            delay={140}
            description="Goals, requests, documents, conversations, and delivery links stay attached from decision to done."
            illustrationClassName="max-w-[20rem]"
            imageSrc="/images/textures/decide-risograph.webp"
            title="Keep context with the work."
            tone="amber"
          >
            <ContextPreview />
          </ShowcaseCard>
        </Box>
      </Box>
    </Container>
  );
};
