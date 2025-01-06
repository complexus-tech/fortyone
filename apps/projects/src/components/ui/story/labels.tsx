import { useLabels } from "@/lib/hooks/labels";
import { Badge, Flex, Tooltip } from "ui";
import { StoryLabel } from "../label";
import { Dot } from "../dot";
import { TagsIcon } from "icons";

export const Labels = ({
  storyLabels = [],
  storyId,
  teamId,
}: {
  storyLabels: string[];
  storyId: string;
  teamId: string;
}) => {
  const { data: allLabels = [] } = useLabels();
  // const labels = allLabels.filter(
  //   (label) => label.teamId === teamId || label.teamId === null,
  // );

  const labels = allLabels.filter((label) => storyLabels.includes(label.id));
  const firstTwoLabels = labels.slice(0, 2);
  const remainingLabels = labels.slice(2);

  return (
    <Flex align="center" gap={1} wrap>
      {firstTwoLabels.map((label) => (
        <StoryLabel key={label.id} {...label} />
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
          <Badge
            rounded="xl"
            color="tertiary"
            className="h-[1.85rem] text-[0.95rem] font-normal"
          >
            <TagsIcon
              className="h-4"
              style={{ color: remainingLabels[0]?.color }}
            />{" "}
            + {remainingLabels.length} label
            {remainingLabels.length > 1 ? "s" : ""}
          </Badge>
        </Tooltip>
      )}
    </Flex>
  );
};
