import { Badge, Flex, Tooltip } from "ui";
import { TagsIcon } from "icons";
import { useLabels } from "@/lib/hooks/labels";
import { useUpdateLabelsMutation } from "@/modules/story/hooks/update-labels-mutation";
import { StoryLabel } from "../label";
import { LabelsMenu } from "./labels-menu";

export const Labels = ({
  storyLabels = [],
  storyId,
  teamId,
}: {
  storyLabels: string[];
  storyId: string;
  teamId: string;
}) => {
  const { mutateAsync } = useUpdateLabelsMutation();
  const { data: allLabels = [] } = useLabels();
  const labels = allLabels.filter((label) => storyLabels.includes(label.id));
  const firstTwoLabels = labels.slice(0, 2);
  const remainingLabels = labels.slice(2);

  const handleUpdateLabels = async (labels: string[] = []) => {
    await mutateAsync({ storyId, labels });
  };

  return (
    <Flex align="center" gap={1} wrap>
      {firstTwoLabels.map((label) => (
        <LabelsMenu key={label.id}>
          <LabelsMenu.Trigger>
            <span>
              <StoryLabel {...label} />
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
                  className="h-[1.85rem] cursor-pointer text-[0.95rem] font-normal"
                  color="tertiary"
                  rounded="xl"
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
