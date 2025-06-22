import { Badge, Flex, Tooltip } from "ui";
import { TagsIcon } from "icons";
import { cn } from "lib";
import type { Label } from "@/types";

export const StoryLabel = ({
  color,
  name,
  isRectangular,
  size = "sm",
}: Label & {
  isRectangular?: boolean;
  size?: "sm" | "md";
}) => {
  return (
    <Tooltip
      title={
        name.length > 12 ? (
          <Flex align="center" gap={1}>
            <TagsIcon className="h-4" style={{ color }} />
            {name}
          </Flex>
        ) : null
      }
    >
      <Badge
        className={cn(
          "h-[1.85rem] cursor-pointer select-none gap-1.5 px-2 text-[0.95rem]",
          {
            "rounded-[0.6rem] px-1.5": isRectangular,
            "h-[2.3rem] rounded-[0.6rem] px-2 text-base": size === "md",
          },
        )}
        color="tertiary"
        rounded={isRectangular ? "md" : "xl"}
        variant="outline"
      >
        <TagsIcon className="h-4" style={{ color }} />
        <span className="inline-block max-w-[12ch] truncate">{name}</span>
      </Badge>
    </Tooltip>
  );
};
