import { cn } from "lib";
import { Box, Flex, Text } from "ui";
import {
  AiIcon,
  CalendarIcon,
  CheckIcon,
  ObjectiveIcon,
  RoadmapIcon,
} from "icons";
import {
  FEATURE_STORY_META_TEXT_CLASS as CARD_META_TEXT_CLASS,
  FEATURE_STORY_SURFACE_CLASS as CARD_SURFACE_CLASS,
  FEATURE_STORY_TEXT_CLASS as CARD_TEXT_CLASS,
  FeatureStoryCard,
  FeatureStorySection,
} from "./feature-story-section";

const NEUTRAL_STATUS_CLASS = cn(
  CARD_META_TEXT_CLASS,
  "bg-accent text-text-secondary shrink-0 rounded-lg px-2.5 py-1 font-semibold",
);

function OutcomePlanCard() {
  return (
    <Box className="flex h-full flex-col gap-3">
      <Box className={cn(CARD_SURFACE_CLASS, "px-4 py-3")}>
        <Flex align="center" className="gap-2.5" justify="between">
          <Flex align="center" className="min-w-0 gap-2.5">
            <Box className="bg-primary/10 flex size-7 shrink-0 items-center justify-center rounded-lg">
              <ObjectiveIcon className="text-primary size-4" strokeWidth={2} />
            </Box>
            <Box className="min-w-0">
              <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
                Company objective
              </Text>
              <Text
                className={cn(
                  CARD_TEXT_CLASS,
                  "text-foreground truncate font-semibold",
                )}
              >
                Increase activation
              </Text>
            </Box>
          </Flex>
          <Text
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-success/10 text-success shrink-0 rounded-lg px-2.5 py-1 font-semibold",
            )}
          >
            On track
          </Text>
        </Flex>
      </Box>

      <Box className={cn(CARD_SURFACE_CLASS, "flex-1 p-4")}>
        <Flex align="center" className="mb-3 gap-2">
          <RoadmapIcon className="text-text-muted size-4" strokeWidth={1.8} />
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            Connected roadmap
          </Text>
        </Flex>
        <Box className="border-border-strong ml-2 border-l-2 border-dashed pl-4">
          <Text className={cn(CARD_TEXT_CLASS, "font-semibold")}>
            Improve onboarding
          </Text>
          <Flex align="center" className="mt-3 gap-2">
            <Box className="border-border-strong flex size-4 shrink-0 items-center justify-center rounded border-2">
              <CheckIcon
                className="text-text-muted size-2.5"
                strokeWidth={2.5}
              />
            </Box>
            <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
              Redesign onboarding flow
            </Text>
          </Flex>
        </Box>
      </Box>

      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "gap-2 px-4 py-2.5")}
      >
        <Box className="bg-primary size-2 rounded-full" />
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          Outcome visible from the work
        </Text>
      </Flex>
    </Box>
  );
}

function CapacityRow({
  fillClassName,
  label,
  value,
}: {
  fillClassName: string;
  label: string;
  value: string;
}) {
  return (
    <Box>
      <Flex align="center" className="mb-2 gap-3" justify="between">
        <Text className={cn(CARD_TEXT_CLASS, "font-semibold")}>{label}</Text>
        <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
          {value}
        </Text>
      </Flex>
      <Box className="bg-surface-muted h-2 overflow-hidden rounded-full">
        <Box className={cn("bg-primary h-full rounded-full", fillClassName)} />
      </Box>
    </Box>
  );
}

function CapacityPlanCard() {
  return (
    <Box className="flex h-full flex-col gap-3">
      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "gap-3 px-4 py-3")}
        justify="between"
      >
        <Flex align="center" className="gap-2.5">
          <Box className="bg-primary/10 flex size-7 items-center justify-center rounded-lg">
            <CalendarIcon className="text-primary size-4" strokeWidth={2} />
          </Box>
          <Text className={cn(CARD_TEXT_CLASS, "font-semibold")}>
            Team capacity
          </Text>
        </Flex>
        <Text className={NEUTRAL_STATUS_CLASS}>This week</Text>
      </Flex>

      <Box className={cn(CARD_SURFACE_CLASS, "grid flex-1 gap-5 p-4")}>
        <CapacityRow fillClassName="w-[68%]" label="Product" value="6h open" />
        <CapacityRow
          fillClassName="w-[84%]"
          label="Engineering"
          value="2h open"
        />
      </Box>

      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "gap-2 px-4 py-2.5")}
      >
        <Box className="bg-success size-2 rounded-full" />
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          Capacity checked before scheduling
        </Text>
      </Flex>
    </Box>
  );
}

function ReviewablePlanCard() {
  return (
    <Box className="flex h-full flex-col gap-3">
      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "gap-2.5 px-4 py-3")}
      >
        <Box className="bg-primary/10 flex size-7 shrink-0 items-center justify-center rounded-lg">
          <AiIcon className="text-primary size-4" />
        </Box>
        <Box className="min-w-0">
          <Text className={cn(CARD_TEXT_CLASS, "font-semibold")}>
            Maya&apos;s plan update
          </Text>
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            Redesign onboarding flow
          </Text>
        </Box>
      </Flex>

      <Box className={cn(CARD_SURFACE_CLASS, "grid flex-1 gap-3 p-4")}>
        <Flex align="center" justify="between">
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            Suggested owner
          </Text>
          <Text className={cn(CARD_TEXT_CLASS, "font-semibold")}>Joseph</Text>
        </Flex>
        <Box className="border-border/80 border-t" />
        <Flex align="center" justify="between">
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            First work block
          </Text>
          <Text className={cn(CARD_TEXT_CLASS, "font-semibold")}>
            Tue 10:30
          </Text>
        </Flex>
      </Box>

      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "gap-2 px-4 py-2.5")}
      >
        <Text className={NEUTRAL_STATUS_CLASS}>Review</Text>
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          No changes applied yet
        </Text>
      </Flex>
    </Box>
  );
}

export const CoreValues = () => {
  return (
    <FeatureStorySection
      heading="Turn priorities into a plan every team can follow."
      id="planning"
    >
      <FeatureStoryCard
        description="Connect planned work to company goals so teams understand what deserves attention first."
        title="Priorities tied to outcomes."
      >
        <OutcomePlanCard />
      </FeatureStoryCard>
      <FeatureStoryCard
        delay={0.1}
        description="Balance new work against availability before the same people become a bottleneck."
        title="Plans that fit capacity."
      >
        <CapacityPlanCard />
      </FeatureStoryCard>
      <FeatureStoryCard
        delay={0.2}
        description="Let Maya suggest an owner and work window, then review the changes before they touch the plan."
        title="AI proposals stay reviewable."
      >
        <ReviewablePlanCard />
      </FeatureStoryCard>
    </FeatureStorySection>
  );
};
