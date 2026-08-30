"use client";

import { Button } from "ui";
import { ObjectiveIcon, OKRIcon, SprintsIcon } from "icons";
import { cn } from "lib";
import { PropertyOption as Option } from "@/components/ui/property-option";
import {
  KeyResultMenu,
  ObjectiveKeyResultMenu,
} from "@/components/ui/story/objective-key-result-menu";
import { SprintsMenu } from "@/components/ui/story/sprints-menu";
import type { DetailedStory } from "../../types";
import { useOptionHotkey } from "./use-option-hotkey";

type TerminologyKey = "keyResultTerm" | "objectiveTerm";

type PlanningOptionsProps = {
  disabled: boolean;
  getTermDisplay: (
    term: TerminologyKey,
    options?: { capitalize?: boolean },
  ) => string;
  isCompact: boolean;
  isNotifications: boolean;
  keyResultId: DetailedStory["keyResultId"];
  keyResultName: string | null;
  objectiveEnabled: boolean;
  objectiveId: DetailedStory["objectiveId"];
  objectiveName: string | null;
  onUpdate: (story: Partial<DetailedStory>) => void;
  sprintId: DetailedStory["sprintId"];
  sprintName?: string;
  sprintsEnabled: boolean;
  teamId: string;
};

export const PlanningOptions = ({
  disabled,
  getTermDisplay,
  isCompact,
  isNotifications,
  keyResultId,
  keyResultName,
  objectiveEnabled,
  objectiveId,
  objectiveName,
  onUpdate,
  sprintId,
  sprintName,
  sprintsEnabled,
  teamId,
}: PlanningOptionsProps) => {
  const objectiveButtonRef = useOptionHotkey(
    "o",
    !disabled && objectiveEnabled,
  );
  const sprintButtonRef = useOptionHotkey("n", !disabled && sprintsEnabled);

  return (
    <>
      {objectiveEnabled ? (
        <>
          <Option
            isCompact={isCompact}
            isNotifications={isNotifications}
            label={getTermDisplay("objectiveTerm", { capitalize: true })}
            value={
              <ObjectiveKeyResultMenu
                align="end"
                keyResultId={keyResultId}
                objectiveId={objectiveId}
                onChange={onUpdate}
                teamId={teamId}
              >
                <Button
                  className="w-fit max-w-[13rem] justify-start font-medium"
                  color="tertiary"
                  disabled={disabled}
                  leftIcon={
                    <ObjectiveIcon
                      className={cn("h-[1.15rem] w-auto shrink-0", {
                        "text-text-muted": !objectiveId,
                      })}
                    />
                  }
                  ref={objectiveButtonRef}
                  size="sm"
                  title={objectiveName ?? undefined}
                  type="button"
                  variant={isCompact ? "solid" : "naked"}
                >
                  <span className="block min-w-0 truncate">
                    {objectiveId
                      ? objectiveName ??
                        getTermDisplay("objectiveTerm", { capitalize: true })
                      : `Add ${getTermDisplay("objectiveTerm")}`}
                  </span>
                </Button>
              </ObjectiveKeyResultMenu>
            }
          />
          {objectiveId ? (
            <Option
              isCompact={isCompact}
              isNotifications={isNotifications}
              label={getTermDisplay("keyResultTerm", { capitalize: true })}
              value={
                <KeyResultMenu
                  align="end"
                  keyResultId={keyResultId}
                  objectiveId={objectiveId}
                  onChange={(nextKeyResultId) => {
                    onUpdate({ keyResultId: nextKeyResultId });
                  }}
                >
                  <Button
                    className="w-fit max-w-[13rem] justify-start font-medium"
                    color="tertiary"
                    disabled={disabled}
                    leftIcon={
                      <OKRIcon
                        className={cn("h-[1.15rem] w-auto shrink-0", {
                          "text-text-muted": !keyResultId,
                        })}
                        strokeWidth={2.4}
                      />
                    }
                    size="sm"
                    title={keyResultName ?? undefined}
                    type="button"
                    variant={isCompact ? "solid" : "naked"}
                  >
                    <span className="block min-w-0 truncate">
                      {keyResultId
                        ? keyResultName ??
                          getTermDisplay("keyResultTerm", { capitalize: true })
                        : `Add ${getTermDisplay("keyResultTerm")}`}
                    </span>
                  </Button>
                </KeyResultMenu>
              }
            />
          ) : null}
        </>
      ) : null}
      {sprintsEnabled ? (
        <Option
          isCompact={isCompact}
          isNotifications={isNotifications}
          label="Sprint"
          value={
            <SprintsMenu>
              <SprintsMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={disabled}
                  leftIcon={
                    <SprintsIcon
                      className={cn("h-5 w-auto", {
                        "text-text-muted": !sprintId,
                      })}
                    />
                  }
                  ref={sprintButtonRef}
                  size="sm"
                  type="button"
                  variant={isCompact ? "solid" : "naked"}
                >
                  <span className="inline-block max-w-[16ch] truncate">
                    {sprintName || "Add sprint"}
                  </span>
                </Button>
              </SprintsMenu.Trigger>
              <SprintsMenu.Items
                align="end"
                setSprintId={(nextSprintId) => {
                  onUpdate({ sprintId: nextSprintId });
                }}
                sprintId={sprintId ?? undefined}
                teamId={teamId}
              />
            </SprintsMenu>
          }
        />
      ) : null}
    </>
  );
};
