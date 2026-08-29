import { Button, Flex, Tooltip } from "ui";
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
      <Button
        className={cn("gap-1 px-2 select-none", {
          "h-[2.3rem] text-base": size === "md",
        })}
        color="tertiary"
        rounded={isRectangular ? "md" : "xl"}
        size="xs"
        type="button"
        variant="outline"
      >
        <TagsIcon className="h-4" style={{ color }} />
        <span className="inline-block max-w-[12ch] truncate">{name}</span>
      </Button>
    </Tooltip>
  );
};
