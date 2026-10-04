import type { ReactNode } from "react";
import { Box, Flex, Text } from "ui";
import { cn } from "lib";

export const SecuritySettingRow = ({
  title,
  description,
  htmlFor,
  layout = "responsive",
  children,
}: {
  title: string;
  description: ReactNode;
  htmlFor?: string;
  layout?: "responsive" | "inline";
  children: ReactNode;
}) => (
  <Flex
    className={cn("gap-4 px-6 py-4", {
      "flex-col items-start md:flex-row md:items-center":
        layout === "responsive",
      "items-center": layout === "inline",
    })}
    justify="between"
  >
    <Box className="min-w-0 flex-1">
      {htmlFor ? (
        <label className="font-medium" htmlFor={htmlFor}>
          {title}
        </label>
      ) : (
        <Text fontWeight="medium">{title}</Text>
      )}
      <Text
        className="mt-1"
        color="muted"
        id={htmlFor ? `${htmlFor}-description` : undefined}
      >
        {description}
      </Text>
    </Box>
    {children}
  </Flex>
);
