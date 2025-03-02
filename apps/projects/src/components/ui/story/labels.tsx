import { Badge, Flex, Tooltip } from "ui";
import { TagsIcon } from "icons";
import { cn } from "lib";
import { useLabels } from "@/lib/hooks/labels";
import { useUpdateLabelsMutation } from "@/modules/story/hooks/update-labels-mutation";
import { useUserRole } from "@/hooks";
import { StoryLabel } from "../label";
import { LabelsMenu } from "./labels-menu";

export const Labels = ({
  storyLabels = [],
  storyId,
  teamId,
  isRectangular,
}: {
  storyLabels: string[];
  storyId: string;
  teamId: string;
  isRectangular?: boolean;
}) => {
  const { mutateAsync } = useUpdateLabelsMutation();
  const { data: allLabels = [] } = useLabels();
  const labels = allLabels.filter((label) => storyLabels.includes(label.id));
  const firstTwoLabels = labels.slice(0, 2);
  const remainingLabels = labels.slice(2);
  const { userRole } = useUserRole();
  const handleUpdateLabels = async (labels: string[] = []) => {
    await mutateAsync({ storyId, labels });
  };

  return (
    <Flex
      align="center"
      className={cn("gap-2", {
        "gap-1.5": isRectangular,
      })}
      wrap
    >
      {firstTwoLabels.map((label) => (
        <LabelsMenu key={label.id}>
          <LabelsMenu.Trigger>
            <span
              className={cn({
                "pointer-events-none cursor-not-allowed": userRole === "guest",
              })}
            >
              <StoryLabel {...label} isRectangular={isRectangular} />
            </span>
          </LabelsMenu.Trigger>
          <LabelsMenu.Items
            labelIds={storyLabels}
            setLabelIds={(labelIds) => {
              handleUpdateLabels(labelIds);
            }}
            teamId={teamId}
          />
        </LabelsMenu>
      ))}
      {remainingLabels.length > 0 && (
        <Tooltip
          title={
            <Flex className="min-w-28" direction="column" gap={2}>
              {remainingLabels.map((label) => (
                <Flex align="center" gap={1} key={label.id}>
                  <TagsIcon className="h-4" style={{ color: label.color }} />
                  {label.name}
                </Flex>
              ))}
            </Flex>
          }
        >
          <span>
            <LabelsMenu>
              <LabelsMenu.Trigger>
                <Badge
                  className={cn(
                    "h-[1.85rem] cursor-pointer text-[0.95rem] font-normal",
                    {
                      "px-1.5": isRectangular,
                      "pointer-events-none cursor-not-allowed":
                        userRole === "guest",
                    },
                  )}
                  color="tertiary"
                  rounded={isRectangular ? "md" : "xl"}
                  variant="outline"
                >
                  <TagsIcon
                    className="h-4"
                    style={{ color: remainingLabels[0]?.color }}
                  />{" "}
                  + {remainingLabels.length} label
                  {remainingLabels.length > 1 ? "s" : ""}
                </Badge>
              </LabelsMenu.Trigger>
              <LabelsMenu.Items
                labelIds={storyLabels}
                setLabelIds={(labelIds) => {
                  handleUpdateLabels(labelIds);
                }}
                teamId={teamId}
              />
            </LabelsMenu>
          </span>
        </Tooltip>
      )}
    </Flex>
  );
};
