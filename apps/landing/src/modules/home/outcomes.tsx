import {
  AnalyticsIcon,
  CheckIcon,
  DashboardIcon,
  OKRIcon,
  ReloadIcon,
  WorkflowIcon,
} from "icons";
import { FeatureGrid } from "@/components/shared/feature-grid";

const capabilities = [
  {
    icon: <DashboardIcon className="h-7 text-dark dark:text-white md:h-9" />,
    title: "Visibility",
    description:
      "Every objective, sprint, and story in one place no more guessing where things stand.",
  },
  {
    icon: <OKRIcon className="h-7 text-dark dark:text-white md:h-9" />,
    title: "Alignment",
    description:
      "OKRs tied directly to daily work so everyone rows in the same direction.",
  },
  {
    icon: <WorkflowIcon className="h-7 text-dark dark:text-white md:h-9" />,
    title: "Efficiency",
    description:
      "Less context switching, fewer status meetings, faster delivery.",
  },
  {
    icon: <ReloadIcon className="h-7 text-dark dark:text-white md:h-9" />,
    title: "Adaptability",
    description: "Plans shift seamlessly when priorities or goals change.",
  },
  {
    icon: <AnalyticsIcon className="h-7 text-dark dark:text-white md:h-9" />,
    title: "Insight",
    description: "Instant reports that show progress, blockers, and impact.",
  },
  {
    icon: <CheckIcon className="h-7 text-dark dark:text-white md:h-9" />,
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
