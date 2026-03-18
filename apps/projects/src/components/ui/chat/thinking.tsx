import type { ReactNode } from "react";
import { Box, Flex, Text } from "ui";
import { cn } from "lib";

export const Thinking = ({
  message = "Maya is thinking",
  icon,
  className,
}: {
  message?: string;
  icon?: ReactNode;
  className?: string;
}) => {
  return (
    <Text
      as="div"
      className={cn("flex items-baseline gap-1.5", className)}
      color="muted"
    >
      {icon && (
        <span className="flex shrink-0 translate-y-[2px] items-center">
          {icon}
        </span>
      )}
      {message}
      <Flex className="gap-0.5">
        <Box className="size-[2.5px] animate-bounce rounded-full bg-current" />
        <Box
          className="size-[2.5px] animate-bounce rounded-full bg-current"
          style={{ animationDelay: "0.1s" }}
        />
        <Box
          className="size-[2.5px] animate-bounce rounded-full bg-current"
          style={{ animationDelay: "0.2s" }}
        />
      </Flex>
    </Text>
  );
};
