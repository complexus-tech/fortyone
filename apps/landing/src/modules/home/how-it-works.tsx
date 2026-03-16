"use client";

import { Text, Box, Flex } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

const viewport = { once: true, amount: 0.3 };
const fadeUp = {
  hidden: { y: 20, opacity: 0 },
  show: { y: 0, opacity: 1, transition: { duration: 0.7, ease: "easeOut" } },
};
const scaleIn = {
  hidden: { scale: 0.96, opacity: 0 },
  show: {
    scale: 1,
    opacity: 1,
    transition: { duration: 0.8, ease: "easeOut" },
  },
};

const cardFadeMaskStyle = {
  WebkitMaskImage:
    "linear-gradient(to bottom, #000 0%, #000 52%, rgba(0, 0, 0, 0.92) 68%, rgba(0, 0, 0, 0.55) 84%, transparent 100%)",
  maskImage:
    "linear-gradient(to bottom, #000 0%, #000 52%, rgba(0, 0, 0, 0.92) 68%, rgba(0, 0, 0, 0.55) 84%, transparent 100%)",
  WebkitMaskRepeat: "no-repeat",
  maskRepeat: "no-repeat",
  WebkitMaskSize: "100% 100%",
  maskSize: "100% 100%",
};

/* ─── Card 01: Task → Goal connection ──────────────────────── */
function TaskGoalCard() {
  return (
    <Box
      className="border-border/40 bg-background relative flex h-full flex-col overflow-hidden rounded-xl border"
      style={cardFadeMaskStyle}
    >
      <Flex align="center" className="gap-1.5 px-3.5 pt-3.5 pb-2.5">
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
      </Flex>

      <Box className="flex flex-1 flex-col px-3.5 pb-3.5">
        <Flex align="center" className="mb-4 gap-2">
          <Box className="bg-text-primary/8 rounded-md px-2 py-1 font-mono text-xs font-semibold tracking-wide opacity-70">
            Q2 OBJECTIVE
          </Box>
          <Text className="text-text-muted text-sm">
            Improve activation rate
          </Text>
        </Flex>

        <Box className="border-border/40 relative ml-3.5 border-l-2 border-dashed pb-1 pl-5">
          <Box className="bg-text-primary/30 absolute top-0 -left-[5px] size-2 rounded-full" />
          <Box className="border-border/50 bg-surface/50 rounded-lg border p-3">
            <Flex align="center" className="mb-2 gap-2">
              <Box className="border-text-primary/30 size-3.5 rounded border-2" />
              <Text className="text-sm font-medium">
                Redesign onboarding flow
              </Text>
            </Flex>
            <Flex align="center" className="gap-2">
              <Box className="bg-text-primary/8 rounded px-1.5 py-0.5 text-xs font-medium opacity-70">
                In Progress
              </Box>
              <Text className="text-text-muted text-xs">Sprint 14</Text>
            </Flex>
          </Box>
        </Box>

        <Box className="border-border/40 relative ml-3.5 border-l-2 border-dashed pb-1 pl-5">
          <Box className="bg-text-primary/20 absolute top-0 -left-[5px] size-2 rounded-full" />
          <Box className="border-border/50 bg-surface/50 rounded-lg border p-3">
            <Flex align="center" className="mb-2 gap-2">
              <Box className="bg-text-primary/10 flex size-3.5 items-center justify-center rounded">
                <Box className="text-text-muted text-xs">✓</Box>
              </Box>
              <Text className="text-text-muted text-sm font-medium line-through decoration-1">
                Add welcome checklist
              </Text>
            </Flex>
            <Flex align="center" className="gap-2">
              <Box className="bg-text-primary/8 text-text-muted rounded px-1.5 py-0.5 text-xs font-medium">
                Done
              </Box>
              <Text className="text-text-muted text-xs">Sprint 13</Text>
            </Flex>
          </Box>
        </Box>

        <Box className="relative ml-3.5">
          <Box className="bg-border/60 absolute top-0 -left-[3px] size-1.5 rounded-full" />
        </Box>
      </Box>
    </Box>
  );
}

/* ─── Card 02: Maya scoping a sprint ───────────────────────── */
function MayaSprintCard() {
  return (
    <Box
      className="border-border/40 bg-background relative flex h-full flex-col overflow-hidden rounded-xl border"
      style={cardFadeMaskStyle}
    >
      <Flex align="center" className="gap-1.5 px-3.5 pt-3.5 pb-2.5">
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
      </Flex>

      <Box className="flex flex-1 flex-col px-3.5 pb-3.5">
        <Flex align="center" className="mb-3 gap-2.5">
          <Box className="bg-text-primary/10 flex size-7 items-center justify-center rounded-full text-sm font-bold">
            M
          </Box>
          <Box>
            <Text className="text-sm font-semibold">Maya</Text>
            <Text className="text-text-muted text-xs">AI Project Manager</Text>
          </Box>
        </Flex>

        <Box className="border-border/30 bg-surface/50 rounded-lg border p-3">
          <Text className="text-text-muted mb-3 text-sm leading-relaxed">
            Based on your team&apos;s capacity, I&apos;d scope these for Sprint
            14:
          </Text>

          <Box className="flex flex-col gap-1.5">
            {["Redesign onboarding flow", "Fix session timeout bug"].map(
              (task, i) => (
                <Flex align="center" className="gap-2" key={task}>
                  <Box className="bg-text-primary/8 flex size-4 items-center justify-center rounded text-xs font-bold opacity-70">
                    {i + 1}
                  </Box>
                  <Text className="text-sm">{task}</Text>
                </Flex>
              ),
            )}
          </Box>
        </Box>

        <Flex className="mt-3 gap-2">
          <Box className="bg-foreground text-background rounded-md px-3 py-1.5 text-sm font-semibold">
            Approve sprint
          </Box>
          <Box className="border-border/50 text-text-muted rounded-md border px-3 py-1.5 text-sm font-medium">
            Adjust
          </Box>
        </Flex>
      </Box>
    </Box>
  );
}

/* ─── Card 03: Gantt-chart roadmap ─────────────────────────── */
const months = ["MAR", "APR", "MAY", "JUN", "JUL"];

const roadmapBars = [
  {
    name: "Onboarding redesign",
    startCol: 1,
    solidCols: 3,
    dashedCols: 1,
    milestones: [
      { col: 2, label: "Wireframes" },
      { col: 3, label: "Ship" },
    ],
  },
  {
    name: "Slack integration",
    startCol: 2,
    solidCols: 2,
    dashedCols: 2,
    milestones: [{ col: 3, label: "Beta" }],
  },
  {
    name: "Analytics dashboard",
    startCol: 3,
    solidCols: 1,
    dashedCols: 2,
    milestones: [{ col: 5, label: "Alpha" }],
  },
];

function RoadmapCard() {
  const totalCols = months.length;

  return (
    <Box
      className="border-border/40 bg-background relative flex h-full flex-col overflow-hidden rounded-xl border"
      style={cardFadeMaskStyle}
    >
      <Flex align="center" className="gap-1.5 px-3.5 pt-3.5 pb-2.5">
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
        <Box className="bg-text-primary/10 size-2.5 rounded-full" />
      </Flex>

      <Box className="flex flex-1 flex-col overflow-hidden px-3.5 pb-3.5">
        {/* Month headers */}
        <Box
          className="mb-3 grid"
          style={{ gridTemplateColumns: `repeat(${totalCols}, 1fr)` }}
        >
          {months.map((m) => (
            <Text
              className="text-text-muted text-center font-mono text-xs tracking-wider"
              key={m}
            >
              {m}
            </Text>
          ))}
        </Box>

        {/* Bars area with grid lines */}
        <Box className="relative flex flex-1 flex-col gap-5">
          {/* Vertical grid lines */}
          <Box className="pointer-events-none absolute inset-0">
            <Box
              className="grid h-full"
              style={{ gridTemplateColumns: `repeat(${totalCols}, 1fr)` }}
            >
              {months.map((m) => (
                <Box
                  className="border-border/15 border-l last:border-r"
                  key={m}
                />
              ))}
            </Box>
          </Box>

          {/* Gantt bars */}
          {roadmapBars.map((bar) => (
            <Box className="relative" key={bar.name}>
              <Text className="mb-2 text-[0.8rem] font-medium">{bar.name}</Text>

              <Box className="relative">
                <Box
                  className="grid items-center"
                  style={{
                    gridTemplateColumns: `repeat(${totalCols}, 1fr)`,
                  }}
                >
                  {/* Solid portion */}
                  <Box
                    className="bg-text-primary/12 relative z-1 h-5 rounded-sm"
                    style={{
                      gridColumn: `${bar.startCol} / span ${bar.solidCols}`,
                    }}
                  />
                  {/* Dashed portion */}
                  {bar.dashedCols > 0 && (
                    <Box
                      className="border-text-primary/20 relative z-1 h-4 rounded-r-sm border border-l-0 border-dashed"
                      style={{
                        gridColumn: `${bar.startCol + bar.solidCols} / span ${bar.dashedCols}`,
                      }}
                    />
                  )}
                </Box>

                {/* Diamond milestones */}
                <Box
                  className="pointer-events-none absolute -bottom-3.5 grid w-full"
                  style={{
                    gridTemplateColumns: `repeat(${totalCols}, 1fr)`,
                  }}
                >
                  {bar.milestones.map((ms) => (
                    <Box
                      className="flex flex-col items-center"
                      key={ms.label}
                      style={{ gridColumn: ms.col }}
                    >
                      <Box className="bg-text-primary/25 size-[5px] rotate-45" />
                      <Text className="text-text-muted relative top-1 mt-1.5 text-xs">
                        {ms.label}
                      </Text>
                    </Box>
                  ))}
                </Box>
              </Box>
            </Box>
          ))}
        </Box>
      </Box>
    </Box>
  );
}

/* ─── Descriptions ─────────────────────────────────────────── */
const steps = [
  {
    number: "01",
    title: "Every task traces to a goal.",
    description:
      "Tasks don't float around a board. Each one links to an objective so your team always knows why the work matters.",
  },
  {
    number: "02",
    title: "Maya scopes sprints for you.",
    description:
      "Tell Maya the outcome you're aiming for. She reads the backlog, weighs capacity, and proposes a sprint you can actually ship.",
  },
  {
    number: "03",
    title: "The full picture, always current.",
    description:
      "Your roadmap updates as work ships. Leaders see what's done, what's next, and what's at risk — without asking.",
  },
];

/* ─── Main Section ─────────────────────────────────────────── */
export const HowItWorks = () => {
  return (
    <Container className="py-16 md:py-28">
      {/* Headline with superscript numbers */}
      <motion.div
        initial="hidden"
        variants={fadeUp}
        viewport={viewport}
        whileInView="show"
      >
        <Text
          as="h2"
          className="mb-14 max-w-3xl text-3xl leading-snug md:mb-20 md:text-5xl md:leading-tight"
        >
          FortyOne connects{" "}
          <sup className="text-text-muted font-mono text-xs font-semibold md:text-sm">
            01
          </sup>{" "}
          every task to a goal, scopes{" "}
          <sup className="text-text-muted font-mono text-xs font-semibold md:text-sm">
            02
          </sup>{" "}
          sprints with Maya, and keeps{" "}
          <sup className="text-text-muted font-mono text-xs font-semibold md:text-sm">
            03
          </sup>{" "}
          progress visible — from plan to shipped.
        </Text>
      </motion.div>

      {/* Three cards */}
      <motion.div
        initial="hidden"
        variants={scaleIn}
        viewport={viewport}
        whileInView="show"
      >
        <Box className="grid grid-cols-1 gap-6 md:grid-cols-3">
          <TaskGoalCard />
          <MayaSprintCard />
          <RoadmapCard />
        </Box>
      </motion.div>

      {/* Numbered descriptions */}
      <Box className="mt-8 grid grid-cols-1 gap-6 md:mt-10 md:grid-cols-3">
        {steps.map((step, i) => (
          <motion.div
            initial="hidden"
            key={step.number}
            transition={{ delay: i * 0.1 }}
            variants={fadeUp}
            viewport={viewport}
            whileInView="show"
          >
            <Box>
              <Text className="text-text-muted mb-2 font-mono text-sm font-semibold">
                {step.number}
              </Text>
              <Text className="mb-2 text-base font-semibold">{step.title}</Text>
              <Text className="text-text-muted text-sm leading-relaxed">
                {step.description}
              </Text>
            </Box>
          </motion.div>
        ))}
      </Box>
    </Container>
  );
};
