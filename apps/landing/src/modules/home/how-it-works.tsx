import { cn } from "lib";
import { Text, Box, Flex } from "ui";
import {
  AiIcon,
  CalendarIcon,
  CommentIcon,
  GitHubIcon,
  GoogleCalendarIcon,
  SettingsIcon,
} from "icons";
import { UnderlinedHandwrittenAccent } from "@/components/ui";
import {
  FEATURE_STORY_META_TEXT_CLASS as CARD_META_TEXT_CLASS,
  FEATURE_STORY_SURFACE_CLASS as CARD_SURFACE_CLASS,
  FEATURE_STORY_TEXT_CLASS as CARD_TEXT_CLASS,
  FeatureStoryCard,
  FeatureStorySection,
} from "./feature-story-section";

function SlackIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zm1.271 0a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313z"
        fill="#E01E5A"
      />
      <path
        d="M8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zm0 1.271a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312z"
        fill="#36C5F0"
      />
      <path
        d="M18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zm-1.27 0a2.528 2.528 0 0 1-2.522 2.521 2.528 2.528 0 0 1-2.521-2.521V2.522A2.528 2.528 0 0 1 15.165 0a2.528 2.528 0 0 1 2.521 2.522v6.312z"
        fill="#2EB67D"
      />
      <path
        d="M15.165 18.956a2.528 2.528 0 0 1 2.521 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.521-2.522v-2.522h2.521zm0-1.27a2.527 2.527 0 0 1-2.521-2.522 2.527 2.527 0 0 1 2.521-2.521h6.313A2.528 2.528 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.521h-6.313z"
        fill="#ECB22E"
      />
    </svg>
  );
}

function IntegrationTile({
  action,
  comfortable = false,
  icon,
  label,
  muted = false,
}: {
  action?: React.ReactNode;
  comfortable?: boolean;
  icon: React.ReactNode;
  label: string;
  muted?: boolean;
}) {
  return (
    <Flex
      align="center"
      className={cn(
        CARD_SURFACE_CLASS,
        "gap-2.5 px-4",
        comfortable ? "py-3" : "py-2",
      )}
      justify="between"
    >
      <Flex align="center" className="min-w-0 gap-2">
        <Box className="flex size-6 shrink-0 items-center justify-center rounded-lg bg-black/4">
          {icon}
        </Box>
        <Text
          className={cn(
            CARD_TEXT_CLASS,
            muted
              ? "text-text-muted font-semibold"
              : "text-foreground font-semibold",
          )}
        >
          {label}
        </Text>
      </Flex>
      {action}
    </Flex>
  );
}

/* ─── Card 01: Request → planned work ─────────────────────── */
type RequestToWorkCardProps = {
  density?: "comfortable" | "default";
};

export function RequestToWorkCard({
  density = "default",
}: RequestToWorkCardProps = {}) {
  const isComfortable = density === "comfortable";

  return (
    <Box className="flex h-full flex-col gap-3">
      <Box
        className={cn(
          CARD_SURFACE_CLASS,
          "px-4 backdrop-blur-sm",
          isComfortable ? "py-4" : "py-3",
        )}
      >
        <Flex align="center" className="gap-3" justify="between">
          <Flex align="center" className="min-w-0 gap-2.5">
            <Box className="bg-primary/10 flex size-7 shrink-0 items-center justify-center rounded-lg">
              <CommentIcon className="text-primary size-4" strokeWidth={2} />
            </Box>
            <Box className="min-w-0">
              <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
                Customer feedback
              </Text>
              <Text
                className={cn(
                  CARD_TEXT_CLASS,
                  "text-foreground truncate font-semibold",
                )}
              >
                Make onboarding easier
              </Text>
            </Box>
          </Flex>
          <Text
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-primary/10 text-primary shrink-0 rounded-lg px-2.5 py-1 font-semibold",
            )}
          >
            12 votes
          </Text>
        </Flex>
      </Box>
      <Box
        className={cn(
          CARD_SURFACE_CLASS,
          "flex-1 px-4",
          isComfortable ? "py-5" : "py-4",
        )}
      >
        <Flex align="center" className="mb-3 gap-3" justify="between">
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            Planned task
          </Text>
          <Text
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-accent text-text-secondary rounded-lg px-2.5 py-1 font-semibold",
              isComfortable && "dark:text-foreground dark:bg-white/12",
            )}
          >
            PRD-142
          </Text>
        </Flex>
        <Box className="border-border-strong ml-3 border-l-2 border-dashed pl-4">
          <Flex align="center" className="gap-2">
            <Box className="border-border-strong size-3.5 rounded border-2" />
            <Text
              className={cn(CARD_TEXT_CLASS, "text-foreground font-medium")}
            >
              Redesign onboarding flow
            </Text>
          </Flex>
          <Text
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-success/10 text-success mt-2 ml-5.5 w-max rounded-lg px-2.5 py-1 font-semibold",
            )}
          >
            Planned
          </Text>
        </Box>
      </Box>
      <Flex
        align="center"
        className={cn(
          CARD_SURFACE_CLASS,
          "gap-2 px-4",
          isComfortable ? "py-3.5" : "py-2.5",
        )}
      >
        <CommentIcon className="text-text-muted size-4 shrink-0" />
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          {isComfortable ? "Feedback linked" : "Original request attached"}
        </Text>
        <Text
          className={cn(
            CARD_META_TEXT_CLASS,
            "bg-accent text-text-secondary ml-auto rounded-lg px-2 py-1 font-semibold",
            isComfortable && "shrink-0 whitespace-nowrap",
          )}
        >
          Goal · Activation
        </Text>
      </Flex>
    </Box>
  );
}

/* ─── Card 02: Integration context picker ─────────────────── */
type IntegrationCardProps = {
  density?: "comfortable" | "default";
};

export function IntegrationCard({
  density = "default",
}: IntegrationCardProps = {}) {
  const isComfortable = density === "comfortable";

  return (
    <Box className="flex h-full flex-col gap-3">
      {/* Command bar */}
      <Flex
        align="center"
        className={cn(
          CARD_SURFACE_CLASS,
          "gap-2 px-4",
          isComfortable ? "py-4" : "py-3",
        )}
      >
        <AiIcon className="text-icon h-4 w-4 shrink-0" />
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          Add context from
        </Text>
        <Text className="text-foreground ml-auto text-[1rem] leading-none font-medium">
          @
        </Text>
      </Flex>
      {/* Integration options */}
      <Box className="grid content-start gap-2">
        <IntegrationTile
          comfortable={isComfortable}
          icon={<GitHubIcon className="size-4.5 shrink-0" />}
          label="GitHub"
        />
        <IntegrationTile
          comfortable={isComfortable}
          icon={<GoogleCalendarIcon className="size-4.5 shrink-0" />}
          label="Google Calendar"
        />
        <IntegrationTile
          action={
            <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
              Connect
            </Text>
          }
          comfortable={isComfortable}
          icon={<SlackIcon className="size-4.5 shrink-0" />}
          label="Slack"
        />
        <IntegrationTile
          action={
            <svg
              className="text-text-muted size-4.5 shrink-0"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              viewBox="0 0 24 24"
            >
              <path
                d="M7 17 17 7M7 7h10v10"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          }
          comfortable={isComfortable}
          icon={
            <SettingsIcon
              className="text-text-muted h-4.5 shrink-0"
              strokeWidth={1.7}
            />
          }
          label="Manage tools"
          muted
        />
      </Box>
    </Box>
  );
}

/* ─── Card 03: Maya assignment plan ───────────────────────── */
type MayaWorkPlanCardProps = {
  density?: "comfortable" | "default";
};

export function MayaWorkPlanCard({
  density = "default",
}: MayaWorkPlanCardProps = {}) {
  const isComfortable = density === "comfortable";

  return (
    <Box className="flex h-full flex-col gap-3">
      <Box
        className={cn(
          CARD_SURFACE_CLASS,
          "px-4",
          isComfortable ? "py-4" : "py-3",
        )}
      >
        <Flex align="center" className="gap-2.5">
          <Box className="bg-primary/10 flex size-7 shrink-0 items-center justify-center rounded-lg">
            <AiIcon className="text-primary size-4" />
          </Box>
          <Box className="min-w-0">
            <Text
              className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
            >
              Maya&apos;s work plan
            </Text>
            <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
              Task: redesign onboarding flow
            </Text>
          </Box>
        </Flex>
      </Box>

      <Box className={cn(CARD_SURFACE_CLASS, "grid gap-2.5 p-4")}>
        <Flex align="center" className="gap-3" justify="between">
          <Box>
            <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
              Suggested owner
            </Text>
            <Text
              className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
            >
              Joseph
            </Text>
          </Box>
          <Text
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-success/10 text-success rounded-lg px-2.5 py-1 font-semibold",
            )}
          >
            Review
          </Text>
        </Flex>
      </Box>

      <Box className="grid grid-cols-2 gap-2">
        <Box
          className={cn(
            CARD_SURFACE_CLASS,
            "px-3",
            isComfortable ? "py-4" : "py-3",
          )}
        >
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            Planned effort
          </Text>
          <Text
            className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
          >
            4 hours
          </Text>
        </Box>
        <Box
          className={cn(
            CARD_SURFACE_CLASS,
            "px-3",
            isComfortable ? "py-4" : "py-3",
          )}
        >
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            First work block
          </Text>
          <Text
            className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
          >
            Tue 10:30
          </Text>
        </Box>
      </Box>

      <Flex
        align="center"
        className={cn(
          CARD_SURFACE_CLASS,
          "mt-auto gap-2 px-4",
          isComfortable ? "py-3.5" : "py-2.5",
        )}
      >
        {isComfortable ? (
          <CalendarIcon
            className="text-warning size-4 shrink-0"
            strokeWidth={2}
          />
        ) : (
          <Box className="bg-warning size-2 rounded-full" />
        )}
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          Calendar and workload checked
        </Text>
      </Flex>
    </Box>
  );
}

/* ─── Main Section ─────────────────────────────────────────── */
export const HowItWorks = () => {
  return (
    <FeatureStorySection
      heading={
        <>
          Decide what matters, plan the work with{" "}
          <UnderlinedHandwrittenAccent tone="danger">
            AI
          </UnderlinedHandwrittenAccent>
          , and deliver it{" "}
          <UnderlinedHandwrittenAccent tone="success">
            together
          </UnderlinedHandwrittenAccent>
          .
        </>
      }
      id="ai-planning"
    >
      <FeatureStoryCard
        description="Bring company goals and customer feedback together, then turn the right decisions into planned work."
        title="Choose the right work."
      >
        <RequestToWorkCard />
      </FeatureStoryCard>
      <FeatureStoryCard
        delay={0.1}
        description="Maya uses workload and calendar availability to suggest an owner and a workable delivery window."
        title="Build a realistic plan."
      >
        <MayaWorkPlanCard />
      </FeatureStoryCard>
      <FeatureStoryCard
        delay={0.2}
        description="Keep the goal, customer request, documents, conversations, and delivery links beside the task so the team can act with full context."
        title="Keep execution connected."
      >
        <IntegrationCard />
      </FeatureStoryCard>
    </FeatureStorySection>
  );
};
