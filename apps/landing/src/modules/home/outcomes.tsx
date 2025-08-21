import { AiIcon, OKRIcon, ReloadIcon, SettingsIcon, SprintsIcon } from "icons";
import { FeatureGrid } from "@/components/shared/feature-grid";

const capabilities = [
  {
    icon: <OKRIcon className="h-8 text-dark dark:text-white" />,
    title: "Visibility",
    description:
      "Every objective, sprint, and story in one place — no more guessing where things stand.",
  },
  {
    icon: <SprintsIcon className="h-8 text-dark dark:text-white" />,
    title: "Alignment",
    description:
      "OKRs tied directly to daily work so everyone rows in the same direction.",
  },
  {
    icon: <AiIcon className="h-8 text-dark dark:text-white" />,
    title: "Efficiency",
    description:
      "Less context switching, fewer status meetings, faster delivery.",
  },
  {
    icon: <ReloadIcon className="h-8 text-dark dark:text-white" />,
    title: "Adaptability",
    description: "Plans shift seamlessly when priorities or goals change.",
  },
  {
    icon: <SettingsIcon className="h-8 text-dark dark:text-white" />,
    title: "Insight",
    description: "Instant reports that show progress, blockers, and impact.",
  },
  {
    icon: <SettingsIcon className="h-8 text-dark dark:text-white" />,
    title: "Confidence",
    description: "Decisions backed by real data, not just gut feeling.",
  },
];

export const CoreOutcomes = () => {
  return (
    <FeatureGrid
      cards={capabilities}
      mainHeading="Outcomes for teams"
      smallHeading="[ What your team gets ]"
    />
  );
};
