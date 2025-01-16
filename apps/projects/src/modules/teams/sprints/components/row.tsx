"use client";
import { Flex, Text, ProgressBar, Box, Badge } from "ui";
import Link from "next/link";
import { ArrowRightIcon, CalendarIcon, SprintsIcon } from "icons";
import { useParams } from "next/navigation";
import { format } from "date-fns";
import { RowWrapper } from "@/components/ui/row-wrapper";
import type { Sprint } from "@/modules/sprints/types";

type SprintStatus = "completed" | "in progress" | "upcoming";

export const SprintRow = ({
  id,
  name,
  startDate,
  endDate,
  stats: { total, completed, cancelled, started, unstarted, backlog },
}: Sprint) => {
  const { teamId } = useParams<{ teamId: string }>();
  let sprintStatus: SprintStatus = "completed";

  if (new Date(startDate) < new Date() && new Date(endDate) > new Date()) {
    sprintStatus = "in progress";
  } else if (new Date(startDate) > new Date()) {
    sprintStatus = "upcoming";
  }

  const progress = Math.round((completed / total) * 100);

  return (
    <RowWrapper className="py-5">
      <Link
        className="flex items-center gap-2"
        href={`/teams/${teamId}/sprints/${id}/stories`}
      >
        <SprintsIcon />
        <Text fontWeight="medium">{name}</Text>
        <Text className="ml-2 flex w-60 items-center gap-1.5" color="muted">
          <CalendarIcon className="relative -top-px" />
          {format(new Date(startDate), "MMM d, yyyy")}
          <ArrowRightIcon className="h-3" />
          {format(new Date(endDate), "MMM d, yyyy")}
        </Text>
      </Link>
      <Flex gap={4}>
        <Box className="">
          <Badge
            className="h-7 px-2 text-[0.9rem] capitalize tracking-wide"
            color={sprintStatus === "in progress" ? "success" : "tertiary"}
          >
            {sprintStatus}
          </Badge>
        </Box>
        <Flex align="center" className="w-24" gap={2}>
          <ProgressBar className="h-1.5 w-12" progress={progress} />
          <Text>{progress}%</Text>
        </Flex>
        <Text className="flex w-24 items-center gap-1.5">
          <span className="font-medium">{completed}</span>
          <Text color="muted">completed</Text>
        </Text>
        <Text className="flex w-20 items-center gap-1.5">
          <span className="font-medium">{total}</span>
          <Text color="muted">total</Text>
        </Text>
      </Flex>
    </RowWrapper>
  );
};
