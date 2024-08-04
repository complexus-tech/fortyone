import { Box, Button, Flex, Menu, Text } from "ui";
import { cn } from "lib";
import { CheckIcon } from "icons";
import { StoryStatusIcon } from "@/components/ui";

export const ObjectiveStatusesMenu = ({
  status,
  isSearchEnabled = true,
  asIcon = true,
}: {
  status: string;
  isSearchEnabled?: boolean;
  asIcon?: boolean;
}) => {
  const statuses = ["Backlog", "In Progress", "Done", "Paused", "Canceled"];
  return <div>test</div>;
};
