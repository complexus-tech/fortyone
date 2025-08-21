import { Flex, Text, Box } from "ui";
import type { ReactNode } from "react";
import {
  AiIcon,
  GitIcon,
  OKRIcon,
  ReloadIcon,
  SettingsIcon,
  SprintsIcon,
} from "icons";
import { Container } from "@/components/ui";

type Capability = {
  icon: ReactNode;
  title: string;
  description: string;
};

const capabilities: Capability[] = [
  {
    icon: <OKRIcon className="h-8 text-dark dark:text-white" />,
    title: "Focus",
    description:
      "Turn big goals into daily progress with objectives and key results.",
  },
  {
    icon: <SprintsIcon className="h-8 text-dark dark:text-white" />,
    title: "Momentum",
    description: "Plan, prioritize, and move fast with stories and sprints.",
  },
  {
    icon: <AiIcon className="h-8 text-dark dark:text-white" />,
    title: "Intelligence",
    description:
      "Work smarter with Maya, your AI assistant for projects and insights.",
  },
  {
    icon: <ReloadIcon className="h-8 text-dark dark:text-white" />,
    title: "Flow",
    description:
      "Stay in sync with real-time updates no refreshing, no delays.",
  },
  {
    icon: <SettingsIcon className="h-8 text-dark dark:text-white" />,
    title: "Identity",
    description:
      "Make the system your own with custom terms for stories, sprints, and goals.",
  },
  {
    icon: <GitIcon className="h-8 text-dark dark:text-white" />,
    title: "Connection",
    description: "Bring everything together with GitHub, emails, and webhooks.",
  },
];

const CapabilityCard = ({ capability }: { capability: Capability }) => {
  return (
    <Box className="group border border-transparent border-l-gray-100 bg-gradient-to-b px-7 py-10 hover:border-gray-100 hover:from-gray-50 dark:border-l-dark-100 dark:hover:border-dark-100 dark:hover:from-dark-300">
      <Flex align="center" className="mb-5" justify="between">
        {capability.icon}
      </Flex>
      <Text as="h3" className="mb-3 text-xl dark:text-white">
        {capability.title}
      </Text>
      <Text className="text-[0.95rem] leading-relaxed opacity-60">
        {capability.description}
      </Text>
    </Box>
  );
};

export const CoreValues = () => {
  return (
    <Container className="pb-28 pt-12">
      <Text className="mb-8 text-sm uppercase tracking-wider opacity-70">
        [ What’s inside ]
      </Text>
      <Text as="h2" className="text-4xl font-semibold md:text-5xl">
        Made for the way you work
      </Text>

      <Box className="mt-24 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {capabilities.map((capability, index) => (
          <CapabilityCard capability={capability} key={index} />
        ))}
      </Box>
    </Container>
  );
};
