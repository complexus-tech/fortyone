export type Plan = {
  name: string;
  highlighted?: boolean;
  limits: {
    members: string;
    fileUploads: string;
    issues: string;
  };
  features: {
    teams: string;
    sso?: boolean;
    emailSupport?: boolean;
    objectives?: string;
    trackOKRs?: boolean;
    unlimitedGuests?: boolean;
    customWorkflows?: boolean;
    customTerminology?: boolean;
    prioritySupport?: boolean;
    privateTeams?: boolean;
    rbac?: boolean;
    unlimitedEverything?: boolean;
    customOnboarding?: boolean;
    onPremise?: boolean;
    dedicatedManager?: boolean;
    volumeDiscounts?: boolean;
    customerFeedback?: boolean;
    feedbackVoting?: boolean;
    publicRoadmap?: boolean;
    strategyMap?: boolean;
    roadmaps?: boolean;
    documents?: boolean;
    calendarPlanning?: boolean;
    myWork?: boolean;
    kanbanAndList?: boolean;
    sprints?: boolean;
    summaryAnalytics?: boolean;
    slackIntegration?: boolean;
    githubIntegration?: boolean;
    figmaIntegration?: boolean;
    googleCalendarIntegration?: boolean;
    mayaMessages?: string;
    aiWorkPlanning?: boolean;
  };
};

export type PlanFeatureKey = keyof Plan["features"];

const CORE_PRODUCT_FEATURES = {
  customerFeedback: true,
  feedbackVoting: true,
  publicRoadmap: true,
  strategyMap: true,
  roadmaps: true,
  documents: true,
  calendarPlanning: true,
  myWork: true,
  kanbanAndList: true,
  sprints: true,
  summaryAnalytics: true,
  slackIntegration: true,
  githubIntegration: true,
  figmaIntegration: true,
  googleCalendarIntegration: true,
} as const satisfies Partial<Plan["features"]>;

export const plans: Plan[] = [
  {
    name: "Hobby",
    limits: {
      members: "Up to 5 members",
      fileUploads: "10MB",
      issues: "Up to 200 tasks",
    },
    features: {
      ...CORE_PRODUCT_FEATURES,
      teams: "1 team",
      objectives: "1 objective",
      trackOKRs: true,
      sso: true,
      emailSupport: true,
      mayaMessages: "15 / month",
    },
  },
  {
    name: "Professional",
    limits: {
      members: "Up to 10 members",
      fileUploads: "25MB",
      issues: "Unlimited",
    },
    features: {
      ...CORE_PRODUCT_FEATURES,
      teams: "Up to 3 teams",
      sso: true,
      rbac: true,
      emailSupport: true,
      objectives: "Up to 20 objectives",
      trackOKRs: true,
      unlimitedGuests: true,
      customWorkflows: true,
      mayaMessages: "100 / month",
      aiWorkPlanning: true,
    },
  },
  {
    name: "Business",
    highlighted: true,
    limits: {
      members: "Unlimited",
      fileUploads: "25MB",
      issues: "Unlimited",
    },
    features: {
      ...CORE_PRODUCT_FEATURES,
      teams: "Unlimited",
      trackOKRs: true,
      sso: true,
      rbac: true,
      emailSupport: true,
      objectives: "Unlimited",
      privateTeams: true,
      customTerminology: true,
      unlimitedEverything: true,
      prioritySupport: true,
      unlimitedGuests: true,
      customWorkflows: true,
      mayaMessages: "500 / month",
      aiWorkPlanning: true,
    },
  },
  {
    name: "Enterprise",
    limits: {
      members: "Unlimited",
      fileUploads: "25MB",
      issues: "Unlimited",
    },
    features: {
      ...CORE_PRODUCT_FEATURES,
      teams: "Unlimited",
      objectives: "Unlimited",
      sso: true,
      rbac: true,
      trackOKRs: true,
      emailSupport: true,
      privateTeams: true,
      unlimitedEverything: true,
      customOnboarding: true,
      onPremise: true,
      dedicatedManager: true,
      prioritySupport: true,
      volumeDiscounts: true,
      unlimitedGuests: true,
      customWorkflows: true,
      customTerminology: true,
      mayaMessages: "Unlimited",
      aiWorkPlanning: true,
    },
  },
];

// Feature labels mapping for display
export const featureLabels = {
  teams: "Teams",
  sso: "Single Sign-On (SSO)",
  emailSupport: "Email support",
  objectives: "Objectives",
  trackOKRs: "Track OKRs",
  unlimitedGuests: "Guests",
  rbac: "Role-based access control",
  privateTeams: "Private teams",
  customWorkflows: "Custom workflows",
  customTerminology: "Custom terminology",
  prioritySupport: "Priority support",
  unlimitedEverything: "Unlimited everything",
  customOnboarding: "Custom onboarding",
  onPremise: "On-premise/Private Cloud Option",
  dedicatedManager: "Dedicated account manager",
  volumeDiscounts: "Volume discounts",
  customerFeedback: "Customer feedback",
  feedbackVoting: "Feedback voting and comments",
  publicRoadmap: "Public roadmap",
  strategyMap: "Strategy map",
  roadmaps: "Roadmaps",
  documents: "Documents",
  calendarPlanning: "Calendar planning",
  myWork: "My Work",
  kanbanAndList: "Kanban and list views",
  sprints: "Sprints",
  summaryAnalytics: "Summary and analytics",
  slackIntegration: "Slack integration",
  githubIntegration: "GitHub integration",
  figmaIntegration: "Figma integration",
  googleCalendarIntegration: "Google Calendar integration",
  mayaMessages: "Maya AI agent messages",
  aiWorkPlanning: "AI work planning",
} satisfies Record<PlanFeatureKey, string>;

// Helper function to generate a list of feature strings from a plan
export const getPlanFeaturesList = (plan: Plan): string[] => {
  const features: string[] = [];

  // Add string features directly
  if (plan.features.teams) features.push(plan.features.teams);
  if (plan.features.objectives) features.push(plan.features.objectives);
  if (plan.features.mayaMessages) {
    features.push(`${plan.features.mayaMessages} Maya AI agent messages`);
  }

  // Add boolean features using the labels
  Object.entries(plan.features).forEach(([key, value]) => {
    if (
      typeof value === "boolean" &&
      value &&
      featureLabels[key as keyof typeof featureLabels]
    ) {
      features.push(featureLabels[key as keyof typeof featureLabels]);
    }
  });

  return features;
};
