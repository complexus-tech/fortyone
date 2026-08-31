import { Text } from "ui";
import {
  getWalkthroughTargetSelector,
  walkthroughTargets,
  type WalkthroughTarget,
} from "@/shared/walkthrough/targets";
import type { WalkthroughStep } from "./walkthrough-provider";

const WORKSPACE_MODULE_TOUR_VERSION = "1.0.0";

interface ModuleWalkthroughTourDefinition {
  key: string;
  navigationDescription: string;
  navigationTarget: WalkthroughTarget;
  navigationTitle: string;
  routes: string[];
  workspaceDescription: string;
  workspaceTitle: string;
}

export interface ModuleWalkthroughTour {
  steps: WalkthroughStep[];
  tourKey: string;
  version: string;
}

const moduleWalkthroughTourDefinitions: ModuleWalkthroughTourDefinition[] = [
  {
    key: "workspace-module-my-work",
    navigationDescription:
      "My work is the quickest route back to the work you own, created, or collaborate on.",
    navigationTarget: walkthroughTargets.myWork,
    navigationTitle: "Find your work from anywhere",
    routes: ["my-work"],
    workspaceDescription:
      "Use this view to review what needs your attention and switch views when a different perspective helps you move it forward.",
    workspaceTitle: "Work from one focused view",
  },
  {
    key: "workspace-module-calendar",
    navigationDescription:
      "Calendar keeps planned work close whenever a commitment needs a real place in your week.",
    navigationTarget: walkthroughTargets.calendar,
    navigationTitle: "Return to your plan",
    routes: ["calendar"],
    workspaceDescription:
      "This workspace brings scheduled work and your calendar together, so you can see whether priorities have enough time behind them.",
    workspaceTitle: "Plan around real time",
  },
  {
    key: "workspace-module-summary",
    navigationDescription:
      "Summary gives you a quick read on how work is moving across the workspace.",
    navigationTarget: walkthroughTargets.summary,
    navigationTitle: "Check the workspace pulse",
    routes: ["summary"],
    workspaceDescription:
      "Use the summary to spot movement, pressure, and areas that need a closer look before opening the underlying work.",
    workspaceTitle: "See the important signals first",
  },
  {
    key: "workspace-module-maya",
    navigationDescription:
      "AI Agent opens Maya whenever you want help planning, clarifying, or preparing the next move.",
    navigationTarget: walkthroughTargets.mayaNavigation,
    navigationTitle: "Open Maya from anywhere",
    routes: ["maya"],
    workspaceDescription:
      "Give Maya a real request and the relevant context. She will explain proposed workspace changes and ask before making them.",
    workspaceTitle: "Work through a real request",
  },
  {
    key: "workspace-module-roadmap",
    navigationDescription:
      "Roadmap keeps individual pieces of work connected to the outcomes they should move.",
    navigationTarget: walkthroughTargets.roadmap,
    navigationTitle: "Keep outcomes visible",
    routes: ["roadmap"],
    workspaceDescription:
      "Use this workspace to shape outcomes, review progress, and keep day-to-day activity connected to a clearer direction.",
    workspaceTitle: "Connect work to direction",
  },
  {
    key: "workspace-module-strategy",
    navigationDescription:
      "Strategy Map gives you a visual route into the outcomes and relationships that shape the plan.",
    navigationTarget: walkthroughTargets.strategy,
    navigationTitle: "Open the strategy map",
    routes: ["strategy"],
    workspaceDescription:
      "Use the map to understand how outcomes support one another and where the strategy needs stronger connections or clearer ownership.",
    workspaceTitle: "See how the strategy connects",
  },
  {
    key: "workspace-module-documents",
    navigationDescription:
      "Documents keeps shared context close to the work instead of scattered across separate tools.",
    navigationTarget: walkthroughTargets.documents,
    navigationTitle: "Return to shared context",
    routes: ["docs"],
    workspaceDescription:
      "Use this workspace for plans, notes, and decisions that people need to find and develop together.",
    workspaceTitle: "Build shared understanding",
  },
  {
    key: "workspace-module-team",
    navigationDescription:
      "Teams is where shared work stays grouped around the people responsible for moving it forward.",
    navigationTarget: walkthroughTargets.teams,
    navigationTitle: "Find a team’s shared work",
    routes: ["teams", "sprints"],
    workspaceDescription:
      "Use the team workspace to move between active work, requests, feedback, objectives, and planning without losing team context.",
    workspaceTitle: "Coordinate in one team context",
  },
];

const routeMatches = (route: string, workspacePath: string) =>
  workspacePath === route || workspacePath.startsWith(`${route}/`);

const getWorkspacePath = (pathname: string, workspaceSlug: string) => {
  if (!workspaceSlug) {
    return null;
  }

  const workspacePrefix = `/${workspaceSlug}`;
  if (
    pathname !== workspacePrefix &&
    !pathname.startsWith(`${workspacePrefix}/`)
  ) {
    return null;
  }

  return pathname.slice(workspacePrefix.length).replace(/^\/+|\/+$/g, "");
};

export const getModuleWalkthroughTour = (
  pathname: string,
  workspaceSlug: string,
): ModuleWalkthroughTour | null => {
  const workspacePath = getWorkspacePath(pathname, workspaceSlug);

  if (!workspacePath) {
    return null;
  }

  const definition = moduleWalkthroughTourDefinitions.find(({ routes }) =>
    routes.some((route) => routeMatches(route, workspacePath)),
  );

  if (!definition) {
    return null;
  }

  return {
    steps: [
      {
        content: <Text color="muted">{definition.navigationDescription}</Text>,
        id: "navigation",
        position: "right",
        target: getWalkthroughTargetSelector(definition.navigationTarget),
        title: definition.navigationTitle,
      },
      {
        content: <Text color="muted">{definition.workspaceDescription}</Text>,
        id: "workspace",
        position: "center",
        target: getWalkthroughTargetSelector(
          walkthroughTargets.workspaceContent,
        ),
        title: definition.workspaceTitle,
      },
    ],
    tourKey: definition.key,
    version: WORKSPACE_MODULE_TOUR_VERSION,
  };
};
