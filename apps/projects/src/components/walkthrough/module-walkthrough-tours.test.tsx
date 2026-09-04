/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getWalkthroughTargetSelector,
  walkthroughTargets,
  type WalkthroughTarget,
} from "@/shared/walkthrough/targets";
import { getModuleWalkthroughTour } from "./module-walkthrough-tours";

type TourExpectation = {
  minimumSteps?: number;
  pathname: string;
  stepTargets: WalkthroughTarget[];
  tourKey: string;
};

const tourExpectations: TourExpectation[] = [
  {
    pathname: "/acme/my-work",
    stepTargets: [
      walkthroughTargets.myWork,
      walkthroughTargets.myWorkTabs,
      walkthroughTargets.myWorkViewControls,
      walkthroughTargets.myWorkFilters,
      walkthroughTargets.myWorkDisplayOptions,
      walkthroughTargets.myWorkContent,
    ],
    tourKey: "workspace-module-my-work",
  },
  {
    pathname: "/acme/calendar",
    stepTargets: [
      walkthroughTargets.calendar,
      walkthroughTargets.calendarDateNavigation,
      walkthroughTargets.calendarView,
      walkthroughTargets.calendarActions,
      walkthroughTargets.calendarGrid,
    ],
    tourKey: "workspace-module-calendar",
  },
  {
    pathname: "/acme/summary",
    stepTargets: [
      walkthroughTargets.summary,
      walkthroughTargets.summaryDateRange,
      walkthroughTargets.summaryOverview,
      walkthroughTargets.summaryHealth,
      walkthroughTargets.summaryMyWork,
      walkthroughTargets.summaryActivityFeed,
    ],
    tourKey: "workspace-module-summary",
  },
  {
    pathname: "/acme/maya",
    stepTargets: [
      walkthroughTargets.mayaNavigation,
      walkthroughTargets.mayaNewChat,
      walkthroughTargets.mayaHeaderActions,
      walkthroughTargets.mayaConversation,
      walkthroughTargets.mayaComposer,
    ],
    tourKey: "workspace-module-maya",
  },
  {
    pathname: "/acme/roadmap",
    stepTargets: [
      walkthroughTargets.roadmap,
      walkthroughTargets.roadmapHeader,
      walkthroughTargets.roadmapLayout,
      walkthroughTargets.roadmapViewOptions,
      walkthroughTargets.roadmapObjectives,
      walkthroughTargets.create,
    ],
    tourKey: "workspace-module-roadmap",
  },
  {
    pathname: "/acme/strategy",
    stepTargets: [
      walkthroughTargets.strategy,
      walkthroughTargets.strategyCanvas,
      walkthroughTargets.strategyCanvasHelp,
      walkthroughTargets.strategyZoom,
      walkthroughTargets.strategyAddPillar,
    ],
    tourKey: "workspace-module-strategy",
  },
  {
    pathname: "/acme/docs/plan-1",
    stepTargets: [
      walkthroughTargets.documents,
      walkthroughTargets.documentsSearch,
      walkthroughTargets.documentsNavigation,
      walkthroughTargets.documentsRecent,
      walkthroughTargets.documentsWorkspace,
    ],
    tourKey: "workspace-module-documents",
  },
  {
    pathname: "/acme/teams/team-1/backlog",
    stepTargets: [
      walkthroughTargets.teams,
      walkthroughTargets.teamNavigation,
      walkthroughTargets.teamSections,
      walkthroughTargets.workspaceContent,
      walkthroughTargets.manageTeams,
    ],
    tourKey: "workspace-module-team",
  },
  {
    minimumSteps: 4,
    pathname: "/acme/sprints",
    stepTargets: [
      walkthroughTargets.sprintsNavigation,
      walkthroughTargets.sprintsHeader,
      walkthroughTargets.sprintsList,
      walkthroughTargets.teams,
    ],
    tourKey: "workspace-module-sprints",
  },
];

describe("getModuleWalkthroughTour", () => {
  it.each(tourExpectations)(
    "maps $pathname to a complete persisted module tour",
    ({ minimumSteps = 5, pathname, stepTargets, tourKey }) => {
      const tour = getModuleWalkthroughTour(pathname, "acme");

      expect(tour).toMatchObject({ tourKey, version: "1.0.0" });
      expect(tour?.steps.map(({ target }) => target)).toEqual(
        stepTargets.map(getWalkthroughTargetSelector),
      );
      expect(tour?.steps.length).toBeGreaterThanOrEqual(minimumSteps);
      expect(new Set(tour?.steps.map(({ id }) => id)).size).toBe(
        tour?.steps.length,
      );
    },
  );

  it.each([
    "/acme/settings",
    "/acme/story/story-1",
    "/another/calendar",
    "/acme-other/calendar",
  ])("does not attach a module tour to %s", (pathname) => {
    expect(getModuleWalkthroughTour(pathname, "acme")).toBeNull();
  });
});
