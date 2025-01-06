import { Label } from "@/types";
import { Badge, Flex, Tooltip } from "ui";
import { TagsIcon } from "icons";

export const StoryLabel = ({ color, name }: Label) => {
  return (
    <Tooltip
      title={
        name?.length > 12 ? (
          <Flex align="center" gap={1}>
            <TagsIcon style={{ color }} className="h-4" />
            {name}
          </Flex>
        ) : null
      }
    >
      <Badge
        className="h-[1.85rem] select-none gap-1.5 px-2 text-[0.95rem] font-normal"
        color="tertiary"
        rounded="xl"
        variant="outline"
      >
        <TagsIcon style={{ color }} className="h-4" />
        <span className="inline-block max-w-[12ch] truncate">{name}</span>
      </Badge>
    </Tooltip>
  );
};
