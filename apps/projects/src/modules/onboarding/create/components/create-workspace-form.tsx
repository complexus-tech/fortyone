"use client";

import type { FormEvent } from "react";
import { useEffect, useRef, useSyncExternalStore } from "react";
import { Text, Button, Flex } from "ui";
import { useProfile } from "@/lib/hooks/profile";
import type { User } from "@/types/user";
import { OnboardingStepper } from "@/components/onboarding/onboarding-stepper";
import { getOnboardingStartUrl } from "@/modules/onboarding/start";
import { useWorkspaceOnboarding } from "../use-workspace-onboarding";
import { WorkspaceDetailsStep } from "./workspace-details-step";
import {
  WorkspaceStartStep,
  WorkspaceWorkStep,
} from "./workspace-onboarding-choices";

const subscribeToClient = () => () => undefined;
const clientSnapshot = () => true;
const serverSnapshot = () => false;
const STEP_TITLES = [
  "Create Workspace",
  "What kind of work?",
  "Make your first move",
] as const;

const WorkspaceForm = ({
  profile,
  callbackUrl,
}: {
  profile: User;
  callbackUrl?: string;
}) => {
  const {
    availability,
    draft,
    error,
    isLoading,
    changeName,
    changeStep,
    updateDraft,
    createWorkspace,
  } = useWorkspaceOnboarding({
    profile,
    onCreated: (workspaceSlug, start) => {
      window.location.href = getOnboardingStartUrl(
        workspaceSlug,
        start,
        callbackUrl,
      );
    },
  });
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (draft.step === 0) changeStep(1);
    else if (draft.step === 1) changeStep(2);
    else void createWorkspace();
  };
  const headingRef = useRef<HTMLHeadingElement>(null);
  const previousStepRef = useRef(draft.step);
  useEffect(() => {
    if (previousStepRef.current === draft.step) return;
    previousStepRef.current = draft.step;
    headingRef.current?.focus({ preventScroll: true });
    headingRef.current?.scrollIntoView({ block: "start" });
  }, [draft.step]);

  return (
    <>
      <h1
        className="mt-8 mb-6 text-4xl font-semibold outline-none"
        ref={headingRef}
        tabIndex={-1}
      >
        {STEP_TITLES[draft.step]}
      </h1>
      <div className="mb-6">
        <OnboardingStepper
          currentStep={draft.step}
          furthestStep={draft.furthestStep}
          onStepChange={
            isLoading
              ? undefined
              : (step) => {
                  if (step === 0 || step === 1 || step === 2) {
                    changeStep(step);
                  }
                }
          }
        />
      </div>
      <form className="space-y-5" onSubmit={handleSubmit}>
        <fieldset className="space-y-5" disabled={isLoading}>
          {draft.step === 0 ? (
            <WorkspaceDetailsStep
              availability={availability}
              draft={draft}
              onChange={updateDraft}
              onNameChange={changeName}
              showFullName={!profile.fullName.trim()}
            />
          ) : null}
          {draft.step === 1 ? (
            <WorkspaceWorkStep draft={draft} onChange={updateDraft} />
          ) : null}
          {draft.step === 2 ? (
            <WorkspaceStartStep draft={draft} onChange={updateDraft} />
          ) : null}
        </fieldset>
        {error ? (
          <Text className="text-danger" role="alert">
            {error}
          </Text>
        ) : null}
        <Flex align="center" className="flex-wrap gap-3" justify="between">
          {draft.step > 0 ? (
            <Button
              align="center"
              className="shrink-0 px-7 md:px-8"
              color="tertiary"
              disabled={isLoading}
              onClick={() => {
                if (draft.step === 2) void createWorkspace("empty");
                else changeStep(0);
              }}
              type="button"
              variant="outline"
            >
              {draft.step === 2 ? "Skip" : "Back"}
            </Button>
          ) : null}
          <Button
            align="center"
            className="ml-auto shrink-0 px-7 md:px-8"
            color="invert"
            disabled={
              (draft.step === 0 &&
                (availability === "checking" ||
                  availability === "unavailable")) ||
              (draft.step === 1 && !draft.workType)
            }
            fullWidth={draft.step === 0}
            loading={isLoading}
            loadingText="Creating…"
            type="submit"
          >
            {
              [
                "Continue to your work",
                "Choose how to start",
                "Create Workspace",
              ][draft.step]
            }
          </Button>
        </Flex>
      </form>
    </>
  );
};

export const CreateWorkspaceForm = ({
  callbackUrl,
}: {
  callbackUrl?: string;
}) => {
  const profileQuery = useProfile();
  const clientReady = useSyncExternalStore(
    subscribeToClient,
    clientSnapshot,
    serverSnapshot,
  );
  if (profileQuery.isError) {
    return (
      <div className="mt-10" role="alert">
        <Text className="mb-4">
          We couldn’t load your account. Please try again.
        </Text>
        <Button
          loading={profileQuery.isFetching}
          onClick={() => void profileQuery.refetch()}
          type="button"
        >
          Try again
        </Button>
      </div>
    );
  }
  if (!clientReady || !profileQuery.data) {
    return (
      <Text className="mt-10" color="muted" role="status">
        Getting your account ready…
      </Text>
    );
  }
  return (
    <WorkspaceForm
      callbackUrl={callbackUrl}
      key={profileQuery.data.id}
      profile={profileQuery.data}
    />
  );
};
