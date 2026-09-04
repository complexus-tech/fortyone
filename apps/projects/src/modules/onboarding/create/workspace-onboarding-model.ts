import { z } from "zod";
import type { WorkspaceWorkType } from "@/lib/actions/create-workspace";

export const TEAM_SIZES = [
  "1-5",
  "6-10",
  "11-50",
  "51-200",
  "201-500",
  "500+",
] as const;

export const WORK_TYPES: {
  value: WorkspaceWorkType;
  label: string;
}[] = [
  {
    value: "product",
    label: "Product & engineering",
  },
  {
    value: "marketing",
    label: "Marketing & creative",
  },
  {
    value: "operations",
    label: "Operations & delivery",
  },
  {
    value: "personal",
    label: "Personal projects",
  },
  {
    value: "general",
    label: "Something else / not sure yet",
  },
];

const stepSchema = z.union([z.literal(0), z.literal(1), z.literal(2)]);
const draftSchema = z.object({
  version: z.literal(1),
  step: stepSchema,
  furthestStep: stepSchema,
  name: z.string().max(80),
  slug: z.string().max(16),
  slugEdited: z.boolean(),
  fullName: z.string().max(120),
  teamSize: z.enum(TEAM_SIZES),
  workType: z
    .enum(["product", "marketing", "operations", "personal", "general"])
    .nullable(),
  start: z.enum(["task", "import", "examples", "empty"]),
});

export type WorkspaceOnboardingDraft = z.infer<typeof draftSchema>;
export type OnboardingStep = WorkspaceOnboardingDraft["step"];
export type WorkspaceStart = WorkspaceOnboardingDraft["start"];

export const getWorkspaceDraftKey = (userId: string) =>
  `fortyone:workspace-onboarding:v1:${userId}`;

export const readWorkspaceDraft = (
  userId: string,
  fullName: string,
): WorkspaceOnboardingDraft => {
  const initial: WorkspaceOnboardingDraft = {
    version: 1,
    step: 0,
    furthestStep: 0,
    name: "",
    slug: "",
    slugEdited: false,
    fullName,
    teamSize: "1-5",
    workType: null,
    start: "task",
  };
  try {
    const stored = sessionStorage.getItem(getWorkspaceDraftKey(userId));
    if (!stored || stored.length > 4000) return initial;
    const parsed = draftSchema.safeParse(JSON.parse(stored));
    if (!parsed.success) return initial;
    const draft = {
      ...parsed.data,
      fullName: fullName || parsed.data.fullName,
    };
    if (getWorkspaceDetailsError(draft)) {
      return { ...draft, step: 0, furthestStep: 0 };
    }
    if (!draft.workType && draft.step === 2) {
      return { ...draft, step: 1, furthestStep: 1 };
    }
    return {
      ...draft,
      furthestStep: Math.max(draft.step, draft.furthestStep) as OnboardingStep,
    };
  } catch {
    // Onboarding remains usable when browser storage is unavailable.
    return initial;
  }
};

export const saveWorkspaceDraft = (
  userId: string,
  draft: WorkspaceOnboardingDraft | null,
) => {
  try {
    const key = getWorkspaceDraftKey(userId);
    if (draft) sessionStorage.setItem(key, JSON.stringify(draft));
    else sessionStorage.removeItem(key);
  } catch {
    // Retain the in-memory draft if browser storage is unavailable.
  }
};

export const formatWorkspaceSlug = (name: string) =>
  name
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 16);

export const isWorkspaceSlugValid = (slug: string) =>
  /^[a-z][a-z0-9-]{2,15}$/.test(slug);

export const getWorkspaceDetailsError = (draft: WorkspaceOnboardingDraft) => {
  if (!draft.fullName.trim()) return "Enter your full name.";
  if (!draft.name.trim()) return "Enter a name for your workspace.";
  if (!isWorkspaceSlugValid(draft.slug)) {
    return "Use 3–16 lowercase letters, numbers, or hyphens for the URL, starting with a letter.";
  }
  return null;
};
