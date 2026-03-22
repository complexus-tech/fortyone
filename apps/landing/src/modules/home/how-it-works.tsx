"use client";

import Image from "next/image";
import { Text, Box, Flex } from "ui";
import { motion } from "framer-motion";
import { AiIcon, GitHubIcon } from "icons";
import { Container } from "@/components/ui";
import meshImage from "../../../public/images/mesh.webp";

const viewport = { once: true, amount: 0.3 };
const fadeUp = {
  hidden: { y: 20, opacity: 0 },
  show: { y: 0, opacity: 1, transition: { duration: 0.7, ease: "easeOut" } },
};

/* ─── Brand icons (inline SVGs) ───────────────────────────── */
function GmailIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M2 6a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6Z"
        fill="#fff"
      />
      <path
        d="M2 6l10 7 10-7"
        stroke="#EA4335"
        strokeWidth="1.8"
        fill="none"
      />
      <rect
        x="2"
        y="4"
        width="20"
        height="16"
        rx="2"
        stroke="#D93025"
        strokeWidth="1.5"
        fill="none"
      />
    </svg>
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

/* ─── Card 01: Task → Goal ────────────────────────────────── */
function TaskGoalCard() {
  return (
    <Box className="flex flex-col gap-3">
      <Box className="bg-background text-text-muted rounded-xl px-4 py-3 text-[0.95rem] shadow-lg">
        Connect task to objective...
      </Box>
      <Box className="bg-background rounded-xl p-4 shadow-lg">
        <Flex align="center" className="mb-3 gap-2">
          <Box className="bg-accent text-text-secondary rounded-md px-2 py-0.5 font-mono text-xs font-semibold tracking-wide">
            Q2 OBJECTIVE
          </Box>
          <Text className="text-text-muted text-[0.95rem]">
            Improve activation rate
          </Text>
        </Flex>
        <Box className="border-border-strong ml-3 border-l-2 border-dashed pl-4">
          <Flex align="center" className="gap-2">
            <Box className="border-border-strong size-3.5 rounded border-2" />
            <Text className="text-foreground text-[0.95rem] font-medium">
              Redesign onboarding flow
            </Text>
          </Flex>
          <Box className="bg-success/10 text-success mt-2 ml-5.5 w-max rounded-full px-2.5 py-0.5 text-xs font-medium">
            In Progress
          </Box>
        </Box>
      </Box>
      {/* Linked source */}
      <Flex
        align="center"
        className="bg-background gap-2 rounded-xl px-4 py-2.5 shadow-lg"
      >
        <GitHubIcon className="size-4 shrink-0" />
        <Text className="text-text-muted text-[0.95rem]">
          Linked to <span className="text-foreground font-medium">#142</span>
        </Text>
        <LinearIcon className="ml-auto size-4 shrink-0" />
        <Text className="text-text-muted text-[0.95rem]">OBJ-7</Text>
      </Flex>
    </Box>
  );
}

/* ─── Card 02: Integration context picker ─────────────────── */
function IntegrationCard() {
  return (
    <Box className="flex flex-col gap-3">
      {/* Command bar */}
      <Flex
        align="center"
        className="bg-background gap-2 rounded-xl px-4 py-3 shadow-lg"
      >
        <AiIcon className="text-icon h-4 w-4 shrink-0" />
        <Text className="text-text-muted text-[0.95rem]">Add context from</Text>
        <Text className="border-border text-text-muted ml-auto rounded-md border px-2 py-0.5 text-[0.95rem]">
          @
        </Text>
      </Flex>
      {/* Integration dropdown */}
      <Box className="bg-background divide-border divide-y rounded-xl shadow-lg">
        <Flex align="center" className="gap-3 px-4 py-3">
          <GitHubIcon className="size-5 shrink-0" />
          <Text className="text-foreground text-[0.95rem] font-medium">
            GitHub
          </Text>
        </Flex>
        <Flex align="center" className="gap-3 px-4 py-3">
          <GmailIcon className="size-5 shrink-0" />
          <Text className="text-foreground text-[0.95rem] font-medium">
            Gmail
          </Text>
        </Flex>
        <Flex align="center" justify="between" className="px-4 py-3">
          <Flex align="center" className="gap-3">
            <SlackIcon className="size-5 shrink-0" />
            <Text className="text-foreground text-[0.95rem] font-medium">
              Slack
            </Text>
          </Flex>
          <Text className="text-text-muted text-[0.95rem]">Connect</Text>
        </Flex>
        <Flex align="center" justify="between" className="px-4 py-3">
          <Flex align="center" className="gap-3">
            <svg
              className="text-text-muted size-5 shrink-0"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.8"
              viewBox="0 0 24 24"
            >
              <path
                d="M12 6V4m0 2a2 2 0 1 0 0 4m0-4a2 2 0 1 1 0 4m-6 8a2 2 0 1 0 0-4m0 4a2 2 0 1 1 0-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 1 0 0-4m0 4a2 2 0 1 1 0-4m0 4v2m0-6V4"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
            <Text className="text-text-muted text-[0.95rem]">
              Manage tools
            </Text>
          </Flex>
          <svg
            className="text-text-muted size-4"
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
        </Flex>
      </Box>
    </Box>
  );
}

/* ─── Card 03: Progress actions ───────────────────────────── */
function ProgressCard() {
  return (
    <Box className="flex flex-col items-end gap-3">
      <Box className="bg-background w-full rounded-xl p-4 shadow-lg">
        <Text className="text-foreground mb-3 text-[0.95rem] font-medium">
          Sprint 14 Progress
        </Text>
        <Box className="bg-surface-muted mb-2.5 h-2.5 w-full overflow-hidden rounded-full">
          <Box className="bg-foreground h-full w-[65%] rounded-full" />
        </Box>
        <Flex justify="between">
          <Text className="text-text-muted text-[0.95rem]">7 of 11 tasks</Text>
          <Text className="text-foreground text-[0.95rem] font-medium">
            65%
          </Text>
        </Flex>
      </Box>
      <Flex className="gap-2">
        <Box className="bg-background-inverse text-foreground-inverse rounded-lg px-3 py-1.5 text-sm font-medium shadow-lg">
          Approve
        </Box>
        <Box className="bg-background text-text-secondary rounded-lg px-3 py-1.5 text-sm font-medium shadow-lg">
          Done
        </Box>
        <Box className="bg-background text-text-muted flex size-8 items-center justify-center rounded-lg text-sm shadow-lg">
          ...
        </Box>
      </Flex>
      {/* Shared to integrations */}
      <Flex
        align="center"
        className="bg-background w-full gap-2.5 rounded-xl px-4 py-2.5 shadow-lg"
      >
        <Text className="text-text-muted text-[0.95rem]">Shared to</Text>
        <Flex className="ml-auto gap-2">
          <SlackIcon className="size-4" />
          <GmailIcon className="size-4" />
        </Flex>
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
    <motion.div
      className="h-full"
      initial="hidden"
      transition={{ delay }}
      variants={fadeUp}
      viewport={viewport}
      whileInView="show"
    >
      <Box className="flex h-full flex-col">
        {/* Mesh gradient area — flex-1 + min-h for tall cards */}
        <Box className="relative flex min-h-[300px] flex-1 items-end overflow-hidden rounded-2xl md:min-h-[350px]">
          <Image
            alt=""
            className="object-cover"
            src={meshImage}
            fill
            quality={100}
          />
          <Box className="relative z-10 w-full p-5">{children}</Box>
        </Box>
        {/* Text below */}
        <Box className="mt-5">
          <Text className="text-foreground mb-2 text-lg font-semibold">
            {title}
          </Text>
          <Text className="text-text-muted">{description}</Text>
        </Box>
      </Box>
    </motion.div>
  );
}

/* ─── Main Section ─────────────────────────────────────────── */
export const HowItWorks = () => {
  return (
    <Container className="py-16 md:py-20">
      {/* Headline */}
      <motion.div
        initial="hidden"
        variants={fadeUp}
        viewport={viewport}
        whileInView="show"
      >
        <Text
          color="gradientDark"
          as="h2"
          className="mb-14 max-w-3xl pb-2 text-3xl md:text-5xl"
        >
          The system that keeps planning, execution, and goals in sync.
        </Text>
      </motion.div>

      {/* Feature cards with mesh backgrounds */}
      <Box className="grid grid-cols-1 gap-6 md:grid-cols-3">
        <FeatureCard
          title="Goals aren't separate from the work."
          description="Tasks, objectives, and delivery live in one system, so progress isn't something you have to reconstruct later."
        >
          <TaskGoalCard />
        </FeatureCard>
        <FeatureCard
          title="Context from every tool."
          description="Gmail, GitHub, Linear, Slack. Maya pulls context so drafts are accurate and follow-ups land at the right time."
          delay={0.1}
        >
          <IntegrationCard />
        </FeatureCard>
        <FeatureCard
          title="Progress stays visible by default."
          description="As work ships, the roadmap reflects it. Teams, managers, and leadership stay aligned on what's done, next, and slipping."
          delay={0.2}
        >
          <ProgressCard />
        </FeatureCard>
      </Box>
    </Container>
  );
};
