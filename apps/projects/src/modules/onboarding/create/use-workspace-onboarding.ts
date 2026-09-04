import { useEffect, useRef, useState } from "react";
import { createWorkspaceAction } from "@/lib/actions/create-workspace";
import { updateProfile } from "@/lib/actions/update-profile";
import { checkWorkspaceAvailability } from "@/lib/queries/check-workspace-availability";
import type { User } from "@/types/user";
import {
  formatWorkspaceSlug,
  getWorkspaceDetailsError,
  isWorkspaceSlugValid,
  readWorkspaceDraft,
  saveWorkspaceDraft,
  type OnboardingStep,
  type WorkspaceOnboardingDraft,
  type WorkspaceStart,
} from "./workspace-onboarding-model";

type Availability = "available" | "unavailable" | "unknown";

export const useWorkspaceOnboarding = ({
  profile,
  onCreated,
}: {
  profile: User;
  onCreated: (workspaceSlug: string, start: WorkspaceStart) => void;
}) => {
  const [draft, setDraft] = useState(() =>
    readWorkspaceDraft(profile.id, profile.fullName.trim()),
  );
  const [availabilityResult, setAvailabilityResult] = useState<{
    slug: string;
    status: Availability;
  } | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const submittingRef = useRef(false);
  const savedProfileRef = useRef<string | null>(null);
  const createdWorkspaceRef = useRef<{
    slug: string;
    start: WorkspaceStart;
  } | null>(null);

  useEffect(() => {
    saveWorkspaceDraft(profile.id, draft);
  }, [draft, profile.id]);

  useEffect(() => {
    if (!isWorkspaceSlugValid(draft.slug)) return;
    let active = true;
    const slug = draft.slug;
    const timeout = window.setTimeout(() => {
      void checkWorkspaceAvailability(slug)
        .then((response) => {
          if (!active) return;
          let status: Availability = "unknown";
          if (!response.error?.message && response.data) {
            status = response.data.available ? "available" : "unavailable";
          }
          setAvailabilityResult({ slug, status });
        })
        .catch(() => {
          if (active) setAvailabilityResult({ slug, status: "unknown" });
        });
    }, 400);
    return () => {
      active = false;
      window.clearTimeout(timeout);
    };
  }, [draft.slug]);

  let availability: Availability | "idle" | "checking" = "idle";
  if (isWorkspaceSlugValid(draft.slug)) {
    availability =
      availabilityResult?.slug === draft.slug
        ? availabilityResult.status
        : "checking";
  }

  const updateDraft = (updates: Partial<WorkspaceOnboardingDraft>) => {
    if (submittingRef.current || createdWorkspaceRef.current) return;
    setError(null);
    setDraft((current) => ({ ...current, ...updates }));
  };

  const changeName = (name: string) => {
    if (submittingRef.current || createdWorkspaceRef.current) return;
    setError(null);
    setDraft((current) => ({
      ...current,
      name,
      ...(!current.slugEdited ? { slug: formatWorkspaceSlug(name) } : {}),
    }));
  };

  const changeStep = (step: OnboardingStep) => {
    if (submittingRef.current || createdWorkspaceRef.current) return;
    if (step > 0) {
      const detailsError = getWorkspaceDetailsError(draft);
      if (detailsError || availability === "unavailable") {
        setError(detailsError ?? "Choose an available workspace URL.");
        setDraft((current) => ({ ...current, step: 0 }));
        return;
      }
      if (availability === "checking") return;
    }
    if (step === 2 && !draft.workType) {
      setError("Choose the kind of work you will manage.");
      setDraft((current) => ({ ...current, step: 1 }));
      return;
    }
    setError(null);
    setDraft((current) => ({
      ...current,
      step,
      furthestStep: Math.max(current.furthestStep, step) as OnboardingStep,
    }));
  };

  const createWorkspace = async (start = draft.start) => {
    if (submittingRef.current) return;
    const detailsError = getWorkspaceDetailsError(draft);
    if (detailsError || !draft.workType) {
      setError(detailsError ?? "Choose the kind of work you will manage.");
      setDraft((current) => ({ ...current, step: detailsError ? 0 : 1 }));
      return;
    }
    submittingRef.current = true;
    setIsLoading(true);
    setError(null);
    try {
      const createdWorkspace = createdWorkspaceRef.current;
      if (createdWorkspace) {
        onCreated(createdWorkspace.slug, createdWorkspace.start);
        return;
      }
      const fullName = profile.fullName.trim() || draft.fullName.trim();
      const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      const profileSignature = JSON.stringify([fullName, timezone]);
      if (savedProfileRef.current !== profileSignature) {
        const result = await updateProfile({ fullName, timezone });
        if (result.error?.message || !result.data) {
          throw new Error(
            result.error?.message ||
              "Your profile could not be saved. Please try again.",
          );
        }
        savedProfileRef.current = profileSignature;
      }
      const response = await createWorkspaceAction({
        name: draft.name.trim(),
        slug: draft.slug,
        teamSize: draft.teamSize,
        workType: draft.workType,
        includeExamples: start === "examples",
      });
      if (response.error?.message || !response.data) {
        throw new Error(
          response.error?.message ||
            "Your workspace could not be created. Please try again.",
        );
      }
      createdWorkspaceRef.current = { slug: response.data.slug, start };
      saveWorkspaceDraft(profile.id, null);
      onCreated(response.data.slug, start);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Something went wrong. Please try again.",
      );
    } finally {
      submittingRef.current = false;
      setIsLoading(false);
    }
  };

  return {
    availability,
    draft,
    error,
    isLoading,
    changeName,
    changeStep,
    updateDraft,
    createWorkspace,
  };
};
