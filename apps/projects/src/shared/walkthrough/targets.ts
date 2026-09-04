export const walkthroughTargets = {
  workspaceSwitcher: "workspace-switcher",
  mayaFloatingAction: "maya-floating-action",
  mayaNavigation: "maya-navigation",
  mayaComposer: "maya-composer",
  mayaConversation: "maya-conversation",
  mayaHeaderActions: "maya-header-actions",
  mayaNewChat: "maya-new-chat",
  calendar: "calendar",
  calendarActions: "calendar-actions",
  calendarDateNavigation: "calendar-date-navigation",
  calendarGrid: "calendar-grid",
  calendarView: "calendar-view",
  roadmap: "roadmap",
  roadmapHeader: "roadmap-header",
  roadmapLayout: "roadmap-layout",
  roadmapObjectives: "roadmap-objectives",
  roadmapViewOptions: "roadmap-view-options",
  strategy: "strategy",
  strategyAddPillar: "strategy-add-pillar",
  strategyCanvas: "strategy-canvas",
  strategyCanvasHelp: "strategy-canvas-help",
  strategyZoom: "strategy-zoom",
  summary: "summary",
  summaryActivityFeed: "summary-activity-feed",
  summaryDateRange: "summary-date-range",
  summaryHealth: "summary-health",
  summaryMyWork: "summary-my-work",
  summaryOverview: "summary-overview",
  documents: "documents",
  documentsNavigation: "documents-navigation",
  documentsRecent: "documents-recent",
  documentsSearch: "documents-search",
  documentsWorkspace: "documents-workspace",
  workspaceContent: "workspace-content",
  notifications: "notifications",
  create: "create",
  myWork: "my-work",
  myWorkContent: "my-work-content",
  myWorkDisplayOptions: "my-work-display-options",
  myWorkFilters: "my-work-filters",
  myWorkTabs: "my-work-tabs",
  myWorkViewControls: "my-work-view-controls",
  teams: "teams",
  teamNavigation: "team-navigation",
  teamSections: "team-sections",
  manageTeams: "manage-teams",
  sprintsNavigation: "sprints-navigation",
  sprintsHeader: "sprints-header",
  sprintsList: "sprints-list",
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
