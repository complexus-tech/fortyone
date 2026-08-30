"use client";

import type { ReactNode } from "react";
import { Avatar, Button, Flex } from "ui";
import { UsersAddIcon } from "icons";
import { cn } from "lib";
import { PropertyOption as Option } from "@/components/ui/property-option";
import { CollaboratorsMenu } from "@/components/ui/story/collaborators-menu";
import { useUpdateCollaboratorsMutation } from "../../hooks/collaboration-mutations";
import type { DetailedStory } from "../../types";
import {
  selectCollaborators,
  type CollaboratorSummary,
} from "./collaborator-selection";

type CollaboratorsOptionProps = {
  assigneeId: string | null;
  collaboratorIds: string[];
  collaborators: DetailedStory["collaborators"];
  disabled: boolean;
  isCompact: boolean;
  isNotifications: boolean;
  members: CollaboratorSummary[];
  storyId: string;
  teamId: string;
};

export const CollaboratorsOption = ({
  assigneeId,
  collaboratorIds,
  collaborators,
  disabled,
  isCompact,
  isNotifications,
  members,
  storyId,
  teamId,
}: CollaboratorsOptionProps) => {
  const { mutate: updateCollaborators } = useUpdateCollaboratorsMutation();
  const selectedCollaborators = selectCollaborators(
    collaboratorIds,
    collaborators,
    members,
  );
  const visibleCollaborators = selectedCollaborators.slice(0, 5);
  const hiddenCollaboratorCount =
    collaboratorIds.length - visibleCollaborators.length;
  const singleCollaborator = selectedCollaborators.at(0);
  let buttonIcon: ReactNode = <UsersAddIcon className="h-[1.15rem] w-auto" />;
  let buttonContent: ReactNode = "Collaborators";

  if (collaboratorIds.length === 1) {
    buttonIcon = (
      <Avatar
        name={singleCollaborator?.fullName || singleCollaborator?.username}
        size="xs"
        src={singleCollaborator?.avatarUrl}
      />
    );
    buttonContent = (
      <span className="max-w-48 truncate">
        {singleCollaborator?.username ||
          singleCollaborator?.fullName ||
          "Collaborator"}
      </span>
    );
  } else if (collaboratorIds.length > 1) {
    buttonIcon = null;
    buttonContent = (
      <Flex className="-space-x-1.5">
        {visibleCollaborators.map((collaborator) => (
          <Avatar
            className="ring-surface ring-1"
            key={collaborator.id}
            name={collaborator.fullName || collaborator.username}
            size="xs"
            src={collaborator.avatarUrl}
          />
        ))}
        {hiddenCollaboratorCount > 0 ? (
          <span className="bg-surface-muted ring-surface flex size-5 items-center justify-center rounded-full text-xs ring-1">
            +{hiddenCollaboratorCount}
          </span>
        ) : null}
      </Flex>
    );
  }

  return (
    <Option
      isCompact={isCompact}
      isNotifications={isNotifications}
      label="Collaborators"
      value={
        <CollaboratorsMenu>
          <CollaboratorsMenu.Trigger>
            <Button
              className={cn("max-w-full font-medium", {
                "text-text-muted": collaboratorIds.length === 0,
              })}
              color="tertiary"
              disabled={disabled}
              leftIcon={buttonIcon}
              size="sm"
              type="button"
              variant={isCompact ? "solid" : "naked"}
            >
              {buttonContent}
            </Button>
          </CollaboratorsMenu.Trigger>
          <CollaboratorsMenu.Items
            assigneeId={assigneeId}
            collaboratorIds={collaboratorIds}
            onCollaboratorsChange={(nextCollaboratorIds) => {
              updateCollaborators({
                storyId,
                collaboratorIds: nextCollaboratorIds,
              });
            }}
            teamId={teamId}
          />
        </CollaboratorsMenu>
      }
    />
  );
};
