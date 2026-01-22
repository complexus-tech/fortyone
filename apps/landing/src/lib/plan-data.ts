type Plan = {
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
  };
};

export const plans: Plan[] = [
  {
    name: "Hobby",
    limits: {
      members: "Up to 5 members",
      fileUploads: "10MB",
      issues: "Up to 200 tasks",
    },
    features: {
      teams: "1 team",
      sso: true,
      emailSupport: true,
    },
  },
  {
    name: "Professional",
    limits: {
      members: "Up to 10 members",
      fileUploads: "Unlimited",
      issues: "Unlimited",
    },
    features: {
      teams: "Up to 3 teams",
      sso: true,
      rbac: true,
      emailSupport: true,
      objectives: "Up to 20 objectives",
      trackOKRs: true,
      unlimitedGuests: true,
      customWorkflows: true,
    },
  },
  {
    name: "Business",
    highlighted: true,
    limits: {
      members: "Unlimited",
      fileUploads: "Unlimited",
      issues: "Unlimited",
    },
    features: {
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
    },
  },
  {
    name: "Enterprise",
    limits: {
      members: "Unlimited",
      fileUploads: "Unlimited",
      issues: "Unlimited",
    },
    features: {
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
};

// Helper function to generate a list of feature strings from a plan
export const getPlanFeaturesList = (plan: Plan): string[] => {
  const features: string[] = [];

  // Add string features directly
  if (plan.features.teams) features.push(plan.features.teams);
  if (plan.features.objectives) features.push(plan.features.objectives);

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
