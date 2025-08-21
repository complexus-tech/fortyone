import {
  AiIcon,
  GitIcon,
  OKRIcon,
  ReloadIcon,
  SettingsIcon,
  SprintsIcon,
} from "icons";
import { FeatureGrid } from "@/components/shared/feature-grid";

const capabilities = [
  {
    icon: <OKRIcon className="h-9 text-dark dark:text-white" />,
    title: "Focus",
    description:
      "Turn big goals into daily progress with objectives and key results.",
  },
  {
    icon: <SprintsIcon className="h-9 text-dark dark:text-white" />,
    title: "Momentum",
    description: "Plan, prioritize, and move fast with stories and sprints.",
  },
  {
    icon: <AiIcon className="h-9 text-dark dark:text-white" />,
    title: "Intelligence",
    description:
      "Work smarter with Maya, your AI assistant for projects and insights.",
  },
  {
    icon: <ReloadIcon className="h-9 text-dark dark:text-white" />,
    title: "Flow",
    description:
      "Stay in sync with real-time updates no refreshing, no delays.",
  },
  {
    icon: <SettingsIcon className="h-9 text-dark dark:text-white" />,
    title: "Identity",
    description:
      "Make the system your own with custom terms for stories, sprints, and goals.",
  },
  {
    icon: <GitIcon className="h-9 text-dark dark:text-white" />,
    title: "Connection",
    description: "Bring everything together with GitHub, emails, and webhooks.",
  },
];

export const CoreValues = () => {
  return (
    <FeatureGrid
      cards={capabilities}
      mainHeading="Made for the way you work"
      smallHeading="[ What's inside ]"
    />
  );
};
