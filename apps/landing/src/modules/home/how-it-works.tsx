import Image from "next/image";
import { cn } from "lib";
import { Text, Box, Button, Flex } from "ui";
import { AiIcon, GitHubIcon, MoreHorizontalIcon, SettingsIcon } from "icons";
import { Container } from "@/components/ui";
import meshImage from "../../../public/images/meshing.webp";

const CARD_TEXT_CLASS = "text-[0.9rem] leading-[1.35]";
const CARD_META_TEXT_CLASS = "text-[0.82rem] leading-[1.25]";
const CARD_SURFACE_CLASS =
  "bg-surface-elevated rounded-xl border border-border/80 shadow-lg shadow-shadow";

/* ─── Brand icons (inline SVGs) ───────────────────────────── */
function DriveIcon({ className }: { className?: string }) {
  return (
    <Image
      alt=""
      className={className}
      height={20}
      src="/integrations/drive.svg"
      width={20}
    />
  );
}

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

function LinearIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M2.414 15.536a.5.5 0 0 1-.048-.604 10 10 0 0 1 6.702-4.792.5.5 0 0 1 .556.32l2.04 5.61a.5.5 0 0 1-.14.547 10 10 0 0 1-4.586 2.319.5.5 0 0 1-.552-.208l-3.972-3.192Z"
        fill="#5E6AD2"
      />
      <path
        d="M3.804 12.196a10 10 0 0 1 16-3.998 10 10 0 0 1 .198 14.002.5.5 0 0 1-.593.106L3.951 12.752a.5.5 0 0 1-.147-.556Z"
        fill="#5E6AD2"
      />
    </svg>
  );
}

function IntegrationTile({
  action,
  icon,
  label,
  muted = false,
}: {
  action?: React.ReactNode;
  icon: React.ReactNode;
  label: string;
  muted?: boolean;
}) {
  return (
    <Flex
      align="center"
      className={cn(CARD_SURFACE_CLASS, "gap-2.5 px-4 py-2")}
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

/* ─── Card 01: Task → Goal ────────────────────────────────── */
function TaskGoalCard() {
  return (
    <Box className="flex h-full flex-col gap-3">
      <Box className={cn(CARD_SURFACE_CLASS, "px-4 py-3 backdrop-blur-sm")}>
        <Flex align="center" className="gap-3" justify="between">
          <Flex align="center" className="min-w-0 gap-2.5">
            <Box className="bg-success/15 flex size-7 shrink-0 items-center justify-center rounded-lg">
              <Box className="bg-success size-2 rounded-full" />
            </Box>
            <Text className={cn(CARD_TEXT_CLASS, "text-text-muted truncate")}>
              Connect task to objective...
            </Text>
          </Flex>
          <Text
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-accent text-text-secondary rounded-lg px-2.5 py-1 font-semibold",
            )}
          >
            Goal
          </Text>
        </Flex>
      </Box>
      <Box className={cn(CARD_SURFACE_CLASS, "flex-1 p-4")}>
        <Flex align="center" className="mb-3 gap-2">
          <Box
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-accent text-text-secondary rounded-lg px-2.5 py-1 font-semibold",
            )}
          >
            Q2 Objective
          </Box>
          <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
            Improve activation rate
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
          <Box
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-success/10 text-success mt-2 ml-5.5 w-max rounded-lg px-2.5 py-1 font-semibold",
            )}
          >
            In Progress
          </Box>
        </Box>
      </Box>
      {/* Linked source */}
      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "gap-2 px-4 py-2.5")}
      >
        <GitHubIcon className="size-4 shrink-0" />
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          Linked to <span className="text-foreground font-medium">#142</span>
        </Text>
        <LinearIcon className="ml-auto size-4 shrink-0" />
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>OBJ-7</Text>
      </Flex>
    </Box>
  );
}

/* ─── Card 02: Integration context picker ─────────────────── */
function IntegrationCard() {
  return (
    <Box className="flex h-full flex-col gap-3">
      {/* Command bar */}
      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "gap-2 px-4 py-3")}
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
          icon={<GitHubIcon className="size-4.5 shrink-0" />}
          label="GitHub"
        />
        <IntegrationTile
          icon={<DriveIcon className="size-4.5 shrink-0" />}
          label="Drive"
        />
        <IntegrationTile
          action={
            <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
              Connect
            </Text>
          }
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

/* ─── Card 03: AI assignment plan ─────────────────────────── */
function AIAssignmentCard() {
  return (
    <Box className="flex h-full flex-col gap-3">
      <Box className={cn(CARD_SURFACE_CLASS, "px-4 py-3")}>
        <Flex align="center" className="gap-2.5">
          <Box className="bg-background-inverse text-foreground-inverse flex size-7 shrink-0 items-center justify-center rounded-lg">
            <AiIcon className="size-4 !text-white dark:!text-black" />
          </Box>
          <Box className="min-w-0">
            <Text
              className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
            >
              AI assignment plan
            </Text>
            <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
              Story: launch customer portal
            </Text>
          </Box>
        </Flex>
      </Box>

      <Box className={cn(CARD_SURFACE_CLASS, "grid gap-2.5 p-4")}>
        <Flex align="center" className="gap-3" justify="between">
          <Box>
            <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
              Best owner
            </Text>
            <Text
              className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
            >
              Priya N.
            </Text>
          </Box>
          <Text
            className={cn(
              CARD_META_TEXT_CLASS,
              "bg-success/10 text-success rounded-lg px-2.5 py-1 font-semibold",
            )}
          >
            92% fit
          </Text>
        </Flex>
        <Box className="bg-surface-muted h-2 overflow-hidden rounded-full">
          <Box className="bg-success h-full w-[82%] rounded-full" />
        </Box>
      </Box>

      <Box className="grid grid-cols-2 gap-2">
        <Box className={cn(CARD_SURFACE_CLASS, "px-3 py-3")}>
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            Estimate
          </Text>
          <Text
            className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
          >
            4 hours
          </Text>
        </Box>
        <Box className={cn(CARD_SURFACE_CLASS, "px-3 py-3")}>
          <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
            Work window
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
        className={cn(CARD_SURFACE_CLASS, "mt-auto gap-2 px-4 py-2.5")}
      >
        <Box className="bg-warning size-2 rounded-full" />
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          Calendar and workload checked
        </Text>
      </Flex>
    </Box>
  );
}

/* ─── Card 04: Progress actions ───────────────────────────── */
function ProgressCard() {
  return (
    <Box className="flex h-full flex-col gap-3">
      <Box className={cn(CARD_SURFACE_CLASS, "w-full p-4")}>
        <Flex align="center" className="mb-3 gap-3" justify="between">
          <Text
            className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
          >
            Sprint 14 Progress
          </Text>
          <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>65%</Text>
        </Flex>
        <Box className="bg-surface-muted mb-4 h-2.5 w-full overflow-hidden rounded-full">
          <Box className="bg-foreground h-full w-[65%] rounded-full" />
        </Box>
        <Box className="grid gap-2">
          {[
            ["API handoff", "Done"],
            ["Billing QA", "In review"],
          ].map(([task, status]) => (
            <Flex
              align="center"
              className="border-border bg-surface-muted/30 rounded-lg border px-3 py-2"
              justify="between"
              key={task}
            >
              <Flex align="center" className="min-w-0 gap-2">
                <Box className="border-success/60 bg-success/15 size-4 shrink-0 rounded-full border" />
                <Text
                  className={cn(
                    CARD_META_TEXT_CLASS,
                    "text-foreground truncate font-medium",
                  )}
                >
                  {task}
                </Text>
              </Flex>
              <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
                {status}
              </Text>
            </Flex>
          ))}
        </Box>
      </Box>

      <Flex className="ml-auto gap-2">
        <Button
          className={cn(CARD_TEXT_CLASS, "shadow-lg")}
          color="invert"
          rounded="lg"
          size="sm"
          type="button"
        >
          Approve
        </Button>
        <Button
          className={cn(CARD_TEXT_CLASS, "shadow-lg")}
          color="tertiary"
          rounded="lg"
          size="sm"
          type="button"
        >
          Done
        </Button>
        <Button
          aria-label="More actions"
          asIcon
          className={cn(CARD_TEXT_CLASS, "text-foreground shadow-lg")}
          color="tertiary"
          rounded="lg"
          size="sm"
          type="button"
        >
          <MoreHorizontalIcon className="h-4 w-auto text-current" />
        </Button>
      </Flex>

      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "mt-auto w-full gap-2.5 px-4 py-2.5")}
      >
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          Shared to
        </Text>
        <Flex className="ml-auto gap-2">
          <SlackIcon className="size-4" />
          <DriveIcon className="size-4" />
        </Flex>
      </Flex>
    </Box>
  );
}

/* ─── Card 05: Capacity planning ──────────────────────────── */
function CapacityCard() {
  const teams = [
    {
      name: "Product",
      value: "72%",
      width: "w-[72%]",
      load: "2 open slots",
    },
    {
      name: "Engineering",
      value: "84%",
      width: "w-[84%]",
      load: "Near limit",
    },
  ];

  return (
    <Box className="flex h-full flex-col gap-3">
      <Box className={cn(CARD_SURFACE_CLASS, "px-4 py-3")}>
        <Flex align="center" className="gap-3" justify="between">
          <Box>
            <Text
              className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
            >
              Team capacity
            </Text>
            <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
              This week
            </Text>
          </Box>
          <Flex className="-space-x-2">
            {[
              {
                alt: "Product lead avatar",
                src: "/images/avatars/product-lead.png",
              },
              {
                alt: "Engineering lead avatar",
                src: "/images/avatars/engineering-lead.png",
              },
              {
                alt: "Operations lead avatar",
                src: "/images/avatars/operations-lead.png",
              },
            ].map((avatar) => (
              <Image
                alt={avatar.alt}
                className="border-background bg-surface-muted size-7 rounded-full border object-cover"
                height={28}
                key={avatar.src}
                src={avatar.src}
                width={28}
              />
            ))}
          </Flex>
        </Flex>
      </Box>

      <Box className="grid gap-2">
        {teams.map((team) => (
          <Box className={cn(CARD_SURFACE_CLASS, "px-4 py-3")} key={team.name}>
            <Flex align="center" className="mb-2 gap-3" justify="between">
              <Box className="min-w-0">
                <Text
                  className={cn(
                    CARD_TEXT_CLASS,
                    "text-foreground truncate font-semibold",
                  )}
                >
                  {team.name}
                </Text>
                <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
                  {team.load}
                </Text>
              </Box>
              <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
                {team.value}
              </Text>
            </Flex>
            <Box className="bg-surface-muted h-2 overflow-hidden rounded-full">
              <Box
                className={cn("bg-foreground h-full rounded-full", team.width)}
              />
            </Box>
          </Box>
        ))}
      </Box>

      <Flex
        align="center"
        className={cn(CARD_SURFACE_CLASS, "mt-auto gap-2.5 px-4 py-2.5")}
      >
        <Box className="bg-success size-2 rounded-full" />
        <Text className={cn(CARD_TEXT_CLASS, "text-text-muted")}>
          AI recommends Product for the next task.
        </Text>
      </Flex>
    </Box>
  );
}

/* ─── Card 06: Review controls ────────────────────────────── */
function ControlCard() {
  return (
    <Box className="flex h-full flex-col gap-3">
      <Box className={cn(CARD_SURFACE_CLASS, "p-4")}>
        <Flex align="start" className="gap-3">
          <Box className="bg-warning/15 flex size-8 shrink-0 items-center justify-center rounded-lg">
            <Box className="bg-warning size-2 rounded-full" />
          </Box>
          <Box>
            <Flex align="center" className="flex-wrap gap-2">
              <Text
                className={cn(CARD_TEXT_CLASS, "text-foreground font-semibold")}
              >
                Review before apply
              </Text>
              <Text
                className={cn(
                  CARD_META_TEXT_CLASS,
                  "bg-warning/10 text-warning rounded-lg px-2 py-0.5 font-semibold",
                )}
              >
                4 changes
              </Text>
            </Flex>
            <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted mt-1")}>
              AI suggested an estimate, owner, and safer start time.
            </Text>
          </Box>
        </Flex>
      </Box>

      <Box className={cn(CARD_SURFACE_CLASS, "grid gap-2.5 p-4")}>
        {[
          ["Estimate", "6 hours"],
          ["Owner", "Priya N."],
          ["Start", "Wed 10:00"],
          ["Task", "Move API handoff"],
        ].map(([label, value]) => (
          <Flex align="center" className="gap-3" justify="between" key={label}>
            <Text className={cn(CARD_META_TEXT_CLASS, "text-text-muted")}>
              {label}
            </Text>
            <Text
              className={cn(
                CARD_META_TEXT_CLASS,
                "text-foreground font-medium",
              )}
            >
              {value}
            </Text>
          </Flex>
        ))}
      </Box>

      <Flex className="mt-auto gap-2">
        <Button
          className={cn(CARD_TEXT_CLASS, "shadow-lg")}
          color="invert"
          rounded="lg"
          size="sm"
          type="button"
        >
          Approve
        </Button>
        <Button
          className={cn(CARD_TEXT_CLASS, "shadow-lg")}
          color="tertiary"
          rounded="lg"
          size="sm"
          type="button"
        >
          Edit
        </Button>
      </Flex>
    </Box>
  );
}

/* ─── Feature card wrapper ─────────────────────────────────── */
function FeatureCard({
  children,
  title,
  description,
  delay = 0,
}: {
  children: React.ReactNode;
  title: string;
  description: string;
  delay?: number;
}) {
  return (
    <Box
      className="h-full"
      data-landing-reveal
      style={{ transitionDelay: `${delay * 1000}ms` }}
    >
      <Box className="flex h-full flex-col">
        <Box className="relative flex h-70 shrink-0 items-end overflow-hidden rounded-2xl md:h-83">
          <Image
            alt=""
            className="object-cover grayscale-100 dark:opacity-40"
            fill
            quality={100}
            sizes="(max-width: 767px) 100vw, (max-width: 1279px) 50vw, 25vw"
            src={meshImage}
          />
          <Box className="relative z-10 w-full p-5">{children}</Box>
        </Box>
        <Box className="mt-5 flex flex-col">
          <Text className="text-foreground mb-2 text-lg font-semibold">
            {title}
          </Text>
          <Text className="text-text-muted">{description}</Text>
        </Box>
      </Box>
    </Box>
  );
}

/* ─── Main Section ─────────────────────────────────────────── */
export const HowItWorks = () => {
  return (
    <Container className="scroll-mt-24 py-16 md:pt-36" id="ai-planning">
      {/* Headline */}
      <Box data-landing-reveal>
        <Text as="h2" className="mb-14 max-w-3xl pb-2 text-3xl md:text-5xl">
          Keep every request connected to the work it shaped.
        </Text>
      </Box>

      {/* Feature cards with mesh backgrounds */}
      <Box className="grid grid-cols-1 gap-6 md:auto-rows-fr md:grid-cols-3">
        <FeatureCard
          description="Turn selected feedback, company goals, and team requests into clear, actionable tasks."
          title="Plan the right work."
        >
          <TaskGoalCard />
        </FeatureCard>
        <FeatureCard
          delay={0.1}
          description="Let AI use workload, estimates, and availability to suggest an owner and time for the work."
          title="Assign with context."
        >
          <AIAssignmentCard />
        </FeatureCard>
        <FeatureCard
          delay={0.2}
          description="Bring conversations, files, and commits into each task so the team can trace every decision."
          title="The original context stays with the task."
        >
          <IntegrationCard />
        </FeatureCard>
      </Box>
    </Container>
  );
};

export const PlatformWorkflow = () => {
  return (
    <Container className="scroll-mt-24 py-16 md:pt-24 md:pb-28" id="roadmaps">
      <Box data-landing-reveal>
        <Text as="h2" className="mb-14 max-w-3xl pb-2 text-3xl md:text-5xl">
          See what is moving, blocked, and at risk.
        </Text>
      </Box>

      <Box className="grid grid-cols-1 gap-6 md:auto-rows-fr md:grid-cols-3">
        <FeatureCard
          description="Keep goals, tasks, and progress connected so the team can see what is done and what needs attention."
          title="Progress without the status chase."
        >
          <ProgressCard />
        </FeatureCard>
        <FeatureCard
          delay={0.1}
          description="Check team workload before new assignments create a bottleneck."
          title="See the bottleneck before assigning."
        >
          <CapacityCard />
        </FeatureCard>
        <FeatureCard
          delay={0.2}
          description="Approve or edit important suggestions before they change the plan."
          title="Maya proposes. Your team approves."
        >
          <ControlCard />
        </FeatureCard>
      </Box>
    </Container>
  );
};
