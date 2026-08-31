/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getWalkthroughTargetSelector,
  walkthroughTargets,
} from "@/shared/walkthrough/targets";
import { getModuleWalkthroughTour } from "./module-walkthrough-tours";

describe("getModuleWalkthroughTour", () => {
  it.each([
    ["/acme/my-work", "workspace-module-my-work", walkthroughTargets.myWork],
    [
      "/acme/calendar",
      "workspace-module-calendar",
      walkthroughTargets.calendar,
    ],
    ["/acme/summary", "workspace-module-summary", walkthroughTargets.summary],
    ["/acme/maya", "workspace-module-maya", walkthroughTargets.mayaNavigation],
    ["/acme/roadmap", "workspace-module-roadmap", walkthroughTargets.roadmap],
    [
      "/acme/strategy",
      "workspace-module-strategy",
      walkthroughTargets.strategy,
    ],
    [
      "/acme/docs/plan-1",
      "workspace-module-documents",
      walkthroughTargets.documents,
    ],
    [
      "/acme/teams/team-1/backlog",
      "workspace-module-team",
      walkthroughTargets.teams,
    ],
    ["/acme/sprints", "workspace-module-team", walkthroughTargets.teams],
  ])("maps %s to its persisted module tour", (pathname, tourKey, target) => {
    const tour = getModuleWalkthroughTour(pathname, "acme");

    expect(tour).toMatchObject({
      tourKey,
      version: "1.0.0",
    });
    expect(tour?.steps).toHaveLength(2);
    expect(tour?.steps[0]?.target).toBe(getWalkthroughTargetSelector(target));
    expect(tour?.steps[1]?.target).toBe(
      getWalkthroughTargetSelector(walkthroughTargets.workspaceContent),
    );
  });

  it.each([
    "/acme/settings",
    "/acme/story/story-1",
    "/another/calendar",
    "/acme-other/calendar",
  ])("does not attach a module tour to %s", (pathname) => {
    expect(getModuleWalkthroughTour(pathname, "acme")).toBeNull();
  });
});
