export const walkthroughTargets = {
  workspaceSwitcher: "workspace-switcher",
  mayaFloatingAction: "maya-floating-action",
  mayaNavigation: "maya-navigation",
  mayaComposer: "maya-composer",
  calendar: "calendar",
  roadmap: "roadmap",
  notifications: "notifications",
  create: "create",
  myWork: "my-work",
  teams: "teams",
  manageTeams: "manage-teams",
  help: "help",
} as const;

export const walkthroughTargetSelectors = {
  createStory:
    '[data-walkthrough-target="create"][data-walkthrough-create-kind="story"]',
} as const;

export type WalkthroughTarget =
  (typeof walkthroughTargets)[keyof typeof walkthroughTargets];

export type WalkthroughTargetSelector =
  | `[data-walkthrough-target="${WalkthroughTarget}"]`
  | (typeof walkthroughTargetSelectors)[keyof typeof walkthroughTargetSelectors];

export const getWalkthroughTargetSelector = (
  target: WalkthroughTarget,
): WalkthroughTargetSelector =>
  `[data-walkthrough-target="${target}"]` as WalkthroughTargetSelector;
