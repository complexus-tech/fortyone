import { Box, Button, Flex, Menu, Text } from "ui";
import { cn } from "lib";
import { CheckIcon } from "icons";
import { StoryStatusIcon } from "../../../components/ui/story-status-icon";

export const ObjectiveStatusesMenu = ({
  isSearchEnabled = true,
  asIcon = true,
}: {
  isSearchEnabled?: boolean;
  asIcon?: boolean;
}) => {
  const statuses = ["Backlog", "In Progress", "Done", "Paused", "Canceled"];
  return <div>Test</div>;
};
