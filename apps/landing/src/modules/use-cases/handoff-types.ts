import type { IntegrationBrandName } from "@/modules/integrations/integration-brand";

export type HandoffIcon =
  | "document"
  | "clock"
  | "link"
  | "team"
  | "goal"
  | "lock"
  | "feedback"
  | "check";

export type HandoffMark = HandoffIcon | IntegrationBrandName | "joseph";

export type ReviewPreview = {
  kind: "review";
  toolbar: string;
  status: string;
  eyebrow: string;
  title: string;
  description: string;
  rows: readonly {
    label: string;
    value: string;
    icon: HandoffIcon;
    tone: "sky" | "sage" | "amber";
  }[];
  decision: { title: string; description: string };
  sources: readonly HandoffMark[];
  note: { title: string; detail: string };
};

export type IntakePreview = {
  kind: "intake";
  source: {
    icon: HandoffMark;
    label: string;
    badge: string;
    author: string;
    authorIcon: HandoffMark;
    message: string;
  };
  task: {
    identifier: string;
    title: string;
    description: string;
    fields: readonly {
      label: string;
      value: string;
      icon: HandoffMark;
    }[];
  };
  sourceCaption: string;
};

export type CapacityPreview = {
  kind: "capacity";
  toolbar: string;
  allocations: readonly [
    { team: string; percent: number },
    { team: string; percent: number },
  ];
  recommendation: string;
  startDay: "Mon" | "Tue" | "Wed" | "Thu" | "Fri";
  proposal: string;
};

export type HandoffPreview = ReviewPreview | IntakePreview | CapacityPreview;

export type HandoffWorkflow = {
  id: string;
  label: string;
  caption: string;
  preview: HandoffPreview;
};

export type HandoffConfig = {
  image:
    | "tasks"
    | "goals"
    | "roadmaps"
    | "customer-feedback"
    | "customer-support"
    | "construction"
    | "ai-planning";
  workflows: readonly [HandoffWorkflow, HandoffWorkflow, HandoffWorkflow];
};
