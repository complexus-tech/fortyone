"use client";

import { useRef } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { Box, Button, Flex } from "ui";
import { TagsIcon } from "icons";
import { cn } from "lib";
import { StoryLabel } from "@/components/ui/label";
import { PropertyOption as Option } from "@/components/ui/property-option";
import { LabelsMenu } from "@/components/ui/story/labels-menu";
import { useLabels } from "@/lib/hooks/labels";
import { useUpdateLabelsMutation } from "../../hooks/update-labels-mutation";
import type { DetailedStory } from "../../types";

type LabelsOptionProps = {
  disabled: boolean;
  isCompact: boolean;
  isNotifications: boolean;
  storyId: string;
  storyLabels: DetailedStory["labels"];
  teamId: string;
};

export const LabelsOption = ({
  disabled,
  isCompact,
  isNotifications,
  storyId,
  storyLabels,
  teamId,
}: LabelsOptionProps) => {
  const { data: allLabels = [] } = useLabels();
  const { mutate: updateLabels } = useUpdateLabelsMutation();
  const labels = allLabels.filter((label) => storyLabels?.includes(label.id));
  const labelsButtonRef = useRef<HTMLButtonElement>(null);
  const emptyLabelsButtonRef = useRef<HTMLButtonElement>(null);

  const handleUpdate = (labelIds: string[] = []) => {
    updateLabels({ storyId, labels: labelIds });
  };

  useHotkeys("l", (event) => {
    event.preventDefault();

    if (!disabled) {
      const target = labels.length > 0 ? labelsButtonRef : emptyLabelsButtonRef;
      target.current?.click();
    }
  });

  return (
    <Option
      className={cn("items-start pt-1", {
        "items-center pt-0": labels.length === 0,
      })}
      isCompact={isCompact}
      isNotifications={isNotifications}
      label="Labels"
      value={
        <Box
          className={cn({
            "md:ml-2.5": !isCompact,
            "md:ml-0": !isCompact && labels.length === 0,
          })}
        >
          {labels.length > 0 ? (
            <Flex align="center" className="gap-1.5" wrap>
              {labels.slice(0, labels.length - 1).map((label) => (
                <LabelsMenu key={label.id}>
                  <LabelsMenu.Trigger>
                    <span
                      className={cn({
                        "pointer-events-none cursor-not-allowed": disabled,
                      })}
                    >
                      <StoryLabel {...label} isRectangular size="sm" />
                    </span>
                  </LabelsMenu.Trigger>
                  <LabelsMenu.Items
                    labelIds={storyLabels ?? []}
                    setLabelIds={handleUpdate}
                    teamId={teamId}
                  />
                </LabelsMenu>
              ))}
              <Flex align="center" gap={1}>
                <LabelsMenu>
                  <LabelsMenu.Trigger>
                    <span
                      className={cn({
                        "pointer-events-none cursor-not-allowed": disabled,
                      })}
                    >
                      <StoryLabel {...labels.at(-1)!} isRectangular size="sm" />
                    </span>
                  </LabelsMenu.Trigger>
                  <LabelsMenu.Items
                    labelIds={storyLabels ?? []}
                    setLabelIds={handleUpdate}
                    teamId={teamId}
                  />
                </LabelsMenu>
                <LabelsMenu>
                  <LabelsMenu.Trigger>
                    <Button
                      asIcon
                      className="m-0"
                      color="tertiary"
                      disabled={disabled}
                      leftIcon={<TagsIcon className="h-4 w-auto" />}
                      ref={labelsButtonRef}
                      rounded="full"
                      size="sm"
                      title="Add labels"
                      type="button"
                      variant={isCompact ? "solid" : "naked"}
                    >
                      <span className="sr-only">Add labels</span>
                    </Button>
                  </LabelsMenu.Trigger>
                  <LabelsMenu.Items
                    labelIds={storyLabels ?? []}
                    setLabelIds={handleUpdate}
                    teamId={teamId}
                  />
                </LabelsMenu>
              </Flex>
            </Flex>
          ) : (
            <LabelsMenu>
              <LabelsMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={disabled}
                  leftIcon={<TagsIcon className="h-[1.15rem] w-auto" />}
                  ref={emptyLabelsButtonRef}
                  size="sm"
                  type="button"
                  variant={isCompact ? "solid" : "naked"}
                >
                  Add labels
                </Button>
              </LabelsMenu.Trigger>
              <LabelsMenu.Items
                labelIds={storyLabels ?? []}
                setLabelIds={handleUpdate}
                teamId={teamId}
              />
            </LabelsMenu>
          )}
        </Box>
      }
    />
  );
};
