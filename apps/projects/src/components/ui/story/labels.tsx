import { useLabels } from "@/lib/hooks/labels";
import { Badge, Flex, Tooltip } from "ui";
import { StoryLabel } from "../label";
import { TagsIcon } from "icons";
import { useUpdateLabelsMutation } from "@/modules/story/hooks/update-labels-mutation";
import { LabelsMenu } from "@/components/ui";

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
            teamId={teamId}
            labelIds={storyLabels}
            setLabelIds={(labelIds) => {
              handleUpdateLabels(labelIds);
            }}
          />
        </LabelsMenu>
      ))}
      {remainingLabels.length > 0 && (
        <Tooltip
          title={
            <Flex className="min-w-28" direction="column" gap={2}>
              {remainingLabels.map((label) => (
                <Flex key={label.id} align="center" gap={1}>
                  <TagsIcon className="h-4" style={{ color: label.color }} />
                  {label.name}
                </Flex>
              ))}
            </Flex>
          }
        >
          <LabelsMenu>
            <LabelsMenu.Trigger>
              <Badge
                rounded="xl"
                color="tertiary"
                className="h-[1.85rem] cursor-pointer text-[0.95rem] font-normal"
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
              teamId={teamId}
              labelIds={storyLabels}
              setLabelIds={(labelIds) => {
                handleUpdateLabels(labelIds);
              }}
            />
          </LabelsMenu>
        </Tooltip>
      )}
    </Flex>
  );
};
