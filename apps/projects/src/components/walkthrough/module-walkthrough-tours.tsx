import { Text } from "ui";
import {
  getWalkthroughTargetSelector,
  walkthroughTargets,
  type WalkthroughTarget,
} from "@/shared/walkthrough/targets";
import type { WalkthroughPosition } from "./walkthrough-position";
import type { WalkthroughStep } from "./walkthrough-provider";

// Keep the original version so users who already completed or skipped a
// module tour are not enrolled again when its guidance is expanded.
const WORKSPACE_MODULE_TOUR_VERSION = "1.0.0";

interface ModuleWalkthroughStepDefinition {
  description: string;
  id: string;
  position: WalkthroughPosition;
  target: WalkthroughTarget;
  title: string;
}

interface ModuleWalkthroughTourDefinition {
  key: string;
  routes: string[];
  steps: ModuleWalkthroughStepDefinition[];
}

export interface ModuleWalkthroughTour {
  steps: WalkthroughStep[];
  tourKey: string;
  version: string;
}

const moduleWalkthroughTourDefinitions: ModuleWalkthroughTourDefinition[] = [
  {
    key: "workspace-module-my-work",
    routes: ["my-work"],
    steps: [
      {
        description:
          "My work is the quickest route back to work you own, created, or collaborate on.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.myWork,
        title: "Find your work from anywhere",
      },
      {
        description:
          "Move between everything, today’s priorities, upcoming work, blockers, assignments, and collaboration without rebuilding a filter.",
        id: "views",
        position: "bottom-end",
        target: walkthroughTargets.myWorkTabs,
        title: "Start with the question you need answered",
      },
      {
        description:
          "Switch between board and list views without changing the work itself. Your preferred layout is remembered.",
        id: "view-controls",
        position: "bottom-end",
        target: walkthroughTargets.myWorkViewControls,
        title: "Choose board or list",
      },
      {
        description:
          "Narrow the view by team, status, priority, people, or dates. Active filters stay visible so you always know what you are looking at.",
        id: "filters",
        position: "bottom-end",
        target: walkthroughTargets.myWorkFilters,
        title: "Focus without losing context",
      },
      {
        description:
          "Choose how work is grouped and ordered, and which supporting details appear on each card or row.",
        id: "display-options",
        position: "bottom-end",
        target: walkthroughTargets.myWorkDisplayOptions,
        title: "Keep the useful details visible",
      },
      {
        description:
          "This is your working surface. Open an item for detail, move it as progress changes, or use a column’s add action to capture the next task.",
        id: "work",
        position: "top",
        target: walkthroughTargets.myWorkContent,
        title: "Move real work forward here",
      },
    ],
  },
  {
    key: "workspace-module-calendar",
    routes: ["calendar"],
    steps: [
      {
        description:
          "Calendar keeps planned work close whenever a commitment needs a real place in your week.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.calendar,
        title: "Return to your plan",
      },
      {
        description:
          "Move through time with the arrows, read the current date range, or jump straight back to today.",
        id: "date-navigation",
        position: "bottom",
        target: walkthroughTargets.calendarDateNavigation,
        title: "Navigate the plan in time",
      },
      {
        description:
          "Use day for detail, week for balance, and month for the broader picture. Display options can include completed work when you need it.",
        id: "view",
        position: "bottom-end",
        target: walkthroughTargets.calendarView,
        title: "Choose the right planning horizon",
      },
      {
        description:
          "Block protected focus time here. You can also schedule assigned work from the Create menu or command bar.",
        id: "actions",
        position: "bottom-end",
        target: walkthroughTargets.calendarActions,
        title: "Turn intentions into reserved time",
      },
      {
        description:
          "Connected meetings, busy time, and scheduled work share this canvas. Open an item for details or move planned work when priorities change.",
        id: "calendar",
        position: "top",
        target: walkthroughTargets.calendarGrid,
        title: "See commitments and capacity together",
      },
    ],
  },
  {
    key: "workspace-module-summary",
    routes: ["summary"],
    steps: [
      {
        description:
          "Summary gives you a quick read on how work is moving across the workspace.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.summary,
        title: "Check the workspace pulse",
      },
      {
        description:
          "Change the reporting window before reading the page. Every signal below updates to the same selected period.",
        id: "date-range",
        position: "bottom-end",
        target: walkthroughTargets.summaryDateRange,
        title: "Choose the period you want to understand",
      },
      {
        description:
          "Closed, overdue, in-progress, created, and assigned counts show the period at a glance. Select any card to open the matching filtered work.",
        id: "overview",
        position: "bottom",
        target: walkthroughTargets.summaryOverview,
        title: "Read the top-level signals first",
      },
      {
        description:
          "Priority, status, and contribution views reveal pressure, stalled work, and how activity is distributed across the team.",
        id: "health",
        position: "top",
        target: walkthroughTargets.summaryHealth,
        title: "Understand what is driving the numbers",
      },
      {
        description:
          "Move between work in progress, due soon, and overdue. Open any task here or jump to My Work for the complete view.",
        id: "my-work",
        position: "top",
        target: walkthroughTargets.summaryMyWork,
        title: "See what needs your attention",
      },
      {
        description:
          "Your activity feed explains the changes behind the numbers, so you can reconnect with recent updates and decisions.",
        id: "activity",
        position: "top",
        target: walkthroughTargets.summaryActivityFeed,
        title: "Understand what changed",
      },
    ],
  },
  {
    key: "workspace-module-maya",
    routes: ["maya"],
    steps: [
      {
        description:
          "AI Agent opens Maya whenever you want help planning, clarifying, or preparing the next move.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.mayaNavigation,
        title: "Open Maya from anywhere",
      },
      {
        description:
          "Start a fresh conversation when the subject changes. Separate chats keep each planning thread easier to revisit.",
        id: "new-chat",
        position: "bottom-start",
        target: walkthroughTargets.mayaNewChat,
        title: "Give each conversation a clear purpose",
      },
      {
        description:
          "History brings past conversations back, while the usage indicator shows how many AI messages remain on your plan.",
        id: "history-and-usage",
        position: "bottom-end",
        target: walkthroughTargets.mayaHeaderActions,
        title: "Return to useful thinking later",
      },
      {
        description:
          "Maya can answer questions about priorities, risks, teams, and delivery. Suggested prompts help you begin with workspace-aware requests.",
        id: "conversation",
        position: "top",
        target: walkthroughTargets.mayaConversation,
        title: "Ask about the work in context",
      },
      {
        description:
          "Write a request, attach supporting files, or use voice. Maya explains proposed workspace changes and asks before making them.",
        id: "composer",
        position: "top-end",
        target: walkthroughTargets.mayaComposer,
        title: "Give Maya the context she needs",
      },
    ],
  },
  {
    key: "workspace-module-roadmap",
    routes: ["roadmap"],
    steps: [
      {
        description:
          "Roadmap keeps individual pieces of work connected to the outcomes they should move.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.roadmap,
        title: "Keep outcomes visible",
      },
      {
        description:
          "The header keeps the current roadmap view clear and surfaces objectives whose forecast needs attention.",
        id: "header",
        position: "bottom-start",
        target: walkthroughTargets.roadmapHeader,
        title: "Start with the roadmap’s current signal",
      },
      {
        description:
          "Use timeline to understand timing, board to compare stages, and list for a compact operational review.",
        id: "layout",
        position: "bottom-end",
        target: walkthroughTargets.roadmapLayout,
        title: "Change perspective without changing the plan",
      },
      {
        description:
          "Board and List add a Customise control here for grouping, ordering, and visible details. Timeline keeps the focus on dates and zoom.",
        id: "view-options",
        position: "bottom-end",
        target: walkthroughTargets.roadmapViewOptions,
        title: "Control the level of detail",
      },
      {
        description:
          "Objectives, key results, owners, health, and dates come together here. Open an objective to inspect or update the plan behind it.",
        id: "objectives",
        position: "top",
        target: walkthroughTargets.roadmapObjectives,
        title: "Connect outcomes to accountable work",
      },
      {
        description:
          "Use Create when a new outcome needs a home, then choose Objective. The same menu is available from anywhere in the app.",
        id: "create",
        position: "bottom-end",
        target: walkthroughTargets.create,
        title: "Add the next outcome when it becomes clear",
      },
    ],
  },
  {
    key: "workspace-module-strategy",
    routes: ["strategy"],
    steps: [
      {
        description:
          "Strategy Map gives you a visual route into the outcomes and relationships that shape the plan.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.strategy,
        title: "Open the strategy map",
      },
      {
        description:
          "Follow the ultimate goal through strategic pillars, objectives, and measurable key results. Select a node to inspect its context and edit it when you have access.",
        id: "map",
        position: "top",
        target: walkthroughTargets.strategyCanvas,
        title: "Trace how outcomes support one another",
      },
      {
        description:
          "Drag cards to organise the map, drag the canvas to pan, and right-click a card for actions without leaving the strategy view.",
        id: "interactions",
        position: "top",
        target: walkthroughTargets.strategyCanvasHelp,
        title: "Arrange and explore the strategy",
      },
      {
        description:
          "Zoom in for detail, zoom out for the full system, or reset the canvas after exploring a large map.",
        id: "zoom",
        position: "bottom-end",
        target: walkthroughTargets.strategyZoom,
        title: "Move comfortably through the map",
      },
      {
        description:
          "Strategic pillars group related outcomes beneath the ultimate goal. Add one when the strategy needs a new durable theme.",
        id: "pillars",
        position: "bottom-end",
        target: walkthroughTargets.strategyAddPillar,
        title: "Give the plan a clear structure",
      },
    ],
  },
  {
    key: "workspace-module-documents",
    routes: ["docs"],
    steps: [
      {
        description:
          "Documents keeps plans, notes, and decisions close to the work instead of scattered across separate tools.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.documents,
        title: "Return to shared context",
      },
      {
        description:
          "Search the workspace’s documents from here or create a blank document without leaving the current context.",
        id: "search",
        position: "right",
        target: walkthroughTargets.documentsSearch,
        title: "Find or capture context quickly",
      },
      {
        description:
          "Move between recent documents, your own work, items shared directly with you, and reusable templates.",
        id: "library",
        position: "right",
        target: walkthroughTargets.documentsNavigation,
        title: "Browse the library by purpose",
      },
      {
        description:
          "Recently updated documents stay close so recurring plans and decisions are easy to continue.",
        id: "recent",
        position: "right",
        target: walkthroughTargets.documentsRecent,
        title: "Pick up where the workspace left off",
      },
      {
        description:
          "The selected library view or open document lives here. Templates give new documents a useful structure, and sharing controls keep access intentional.",
        id: "workspace",
        position: "left",
        target: walkthroughTargets.documentsWorkspace,
        title: "Develop shared understanding together",
      },
    ],
  },
  {
    key: "workspace-module-team",
    routes: ["teams"],
    steps: [
      {
        description:
          "Your Teams keeps shared work grouped around the people responsible for moving it forward.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.teams,
        title: "Find every team in the sidebar",
      },
      {
        description:
          "Expand a team to enter its working context. The active team stays expanded so you always know which group you are working with.",
        id: "team",
        position: "right",
        target: walkthroughTargets.teamNavigation,
        title: "Keep the current team visible",
      },
      {
        description:
          "Move between intake, feedback, tasks, objectives, and sprints without losing the team context. Available sections reflect the team’s setup.",
        id: "sections",
        position: "right",
        target: walkthroughTargets.teamSections,
        title: "Follow work through the team’s workflow",
      },
      {
        description:
          "The selected team section opens here. Its controls and views are scoped to this team, so changes stay in the right context.",
        id: "workspace",
        position: "center",
        target: walkthroughTargets.workspaceContent,
        title: "Work inside one clear team context",
      },
      {
        description:
          "Use this menu to find more teams, pin the ones you use most, or manage membership and team settings when your role allows it.",
        id: "manage",
        position: "right",
        target: walkthroughTargets.manageTeams,
        title: "Keep the sidebar useful as teams grow",
      },
    ],
  },
  {
    key: "workspace-module-sprints",
    routes: ["sprints"],
    steps: [
      {
        description:
          "Active Sprints brings every currently running team cycle into one workspace-level view.",
        id: "navigation",
        position: "right",
        target: walkthroughTargets.sprintsNavigation,
        title: "Return to active delivery cycles",
      },
      {
        description:
          "This page is the cross-team entry point for live sprints, rather than a backlog of every historical cycle.",
        id: "header",
        position: "bottom-start",
        target: walkthroughTargets.sprintsHeader,
        title: "Focus on what is running now",
      },
      {
        description:
          "Each row shows a sprint’s team, dates, and current state. Open one to review the work and progress inside that cycle.",
        id: "sprints",
        position: "top",
        target: walkthroughTargets.sprintsList,
        title: "Move from portfolio view to sprint detail",
      },
      {
        description:
          "Sprints belong to teams. Expand a team here to find its sprint history, objectives, tasks, feedback, and intake together.",
        id: "teams",
        position: "right",
        target: walkthroughTargets.teams,
        title: "Follow a sprint back to its team",
      },
    ],
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
    steps: definition.steps.map((step) => ({
      content: <Text color="muted">{step.description}</Text>,
      id: step.id,
      position: step.position,
      target: getWalkthroughTargetSelector(step.target),
      title: step.title,
    })),
    tourKey: definition.key,
    version: WORKSPACE_MODULE_TOUR_VERSION,
  };
};
