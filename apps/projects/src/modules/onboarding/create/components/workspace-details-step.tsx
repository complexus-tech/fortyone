import { Input, Text, Flex } from "ui";
import { CloseIcon } from "icons";
import type { WorkspaceOnboardingDraft } from "../workspace-onboarding-model";
import type { useWorkspaceOnboarding } from "../use-workspace-onboarding";

const isFortyOneApp = process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";
const AVAILABILITY_HELP = {
  unavailable: "This URL is already taken. Please try a different one.",
  checking: "Checking availability…",
  unknown:
    "We couldn’t check availability. You can continue and try creating the workspace.",
  available: "Pick a simple, memorable URL for your workspace",
  idle: "Pick a simple, memorable URL for your workspace",
} as const;

type WorkspaceDetailsStepProps = {
  availability: ReturnType<typeof useWorkspaceOnboarding>["availability"];
  draft: Pick<WorkspaceOnboardingDraft, "name" | "slug" | "fullName">;
  onChange: (updates: Partial<WorkspaceOnboardingDraft>) => void;
  onNameChange: (name: string) => void;
  showFullName: boolean;
};

export const WorkspaceDetailsStep = ({
  availability,
  draft,
  onChange,
  onNameChange,
  showFullName,
}: WorkspaceDetailsStepProps) => (
  <>
    {showFullName ? (
      <Input
        autoComplete="name"
        className="rounded-lg"
        label="Your full name"
        maxLength={120}
        name="fullName"
        onChange={(event) => {
          onChange({ fullName: event.target.value });
        }}
        placeholder="Enter your full name"
        required
        value={draft.fullName}
      />
    ) : null}
    <Input
      className="rounded-lg"
      label="Your Workspace"
      maxLength={80}
      name="name"
      onChange={(event) => {
        onNameChange(event.target.value);
      }}
      placeholder="Enter workspace name"
      required
      value={draft.name}
    />
    <Input
      className="rounded-lg"
      hasError={availability === "unavailable"}
      helpText={AVAILABILITY_HELP[availability]}
      label="Workspace URL"
      maxLength={16}
      minLength={3}
      name="slug"
      onChange={(event) => {
        onChange({ slug: event.target.value.toLowerCase(), slugEdited: true });
      }}
      pattern="^[a-z][a-z0-9-]*$"
      required
      rightIcon={
        <Flex align="center" gap={2}>
          {isFortyOneApp ? <Text>.fortyone.app</Text> : null}
          {availability === "unavailable" ? (
            <Flex
              align="center"
              className="bg-danger size-5 rounded-full"
              justify="center"
            >
              <CloseIcon
                className="h-3 text-white dark:text-white"
                strokeWidth={3}
              />
            </Flex>
          ) : null}
        </Flex>
      }
      value={draft.slug}
    />
  </>
);
