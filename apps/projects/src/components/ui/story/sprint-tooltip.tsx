import { Flex, Text, Box, Badge } from "ui";
import { cn } from "lib";
import { CalendarIcon, SprintsIcon } from "icons";
import { format, differenceInDays } from "date-fns";
import type { Sprint } from "@/modules/sprints/types";

export const sprintTooltip = (selectedSprint: Sprint | undefined) => {
  const isCompleted = new Date(selectedSprint?.endDate!) < new Date();
  const inProgress =
    new Date(selectedSprint?.startDate!) < new Date() &&
    new Date(selectedSprint?.endDate!) > new Date();
  const daysLeft = differenceInDays(
    new Date(selectedSprint?.endDate!),
    new Date(),
  );
  const isPanned = new Date(selectedSprint?.startDate!) > new Date();

  const getBadgeColor = () => {
    if (isCompleted || isPanned) {
      return "tertiary";
    }
    if (inProgress && daysLeft < 5) {
      return "primary";
    }
    if (inProgress && daysLeft < 8) {
      return "warning";
    }
    return "info";
  };

  const getBadgeText = () => {
    if (isCompleted) {
      return "Completed";
    }
    if (isPanned) {
      return "Planned";
    }
    if (inProgress && daysLeft < 5) {
      return `${daysLeft} days left`;
    }
    if (inProgress && daysLeft < 8) {
      return "Ending in a week";
    }
    return "In progress";
  };

  return (
    <Box>
      <Text className="flex items-center gap-1" fontSize="md">
        <SprintsIcon className="h-5 w-auto shrink-0" />
        {selectedSprint?.name}
      </Text>
      <Flex align="center" className="mb-3 mt-4" gap={6} justify="between">
        <Text className="flex items-center gap-1" fontSize="md">
          <CalendarIcon
            className={cn("h-5 w-auto", {
              "text-primary dark:text-primary": getBadgeColor() === "primary",
              "text-warning dark:text-warning": getBadgeColor() === "warning",
              "text-info dark:text-info": getBadgeColor() === "info",
            })}
          />{" "}
          {format(new Date(selectedSprint?.startDate!), "MMM dd")} -{" "}
          {format(new Date(selectedSprint?.endDate!), "MMM dd")}
        </Text>
        <Badge
          className="h-7 text-[0.95rem] font-medium capitalize"
          color={getBadgeColor()}
          rounded="sm"
        >
          {getBadgeText()}
        </Badge>
      </Flex>
      {selectedSprint?.goal ? <>
          <Text fontSize="md">Sprint Goal:</Text>
          <Text className="mt-1 line-clamp-4" color="muted" fontSize="md">
            {selectedSprint.goal}
          </Text>
        </> : null}
    </Box>
  );
};
