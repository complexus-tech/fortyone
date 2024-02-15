import { Flex, Text, Tooltip } from "ui";
import Link from "next/link";
import { Calendar, CalendarCheck2, Hexagon } from "lucide-react";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { AssigneesMenu } from "@/components/ui/issue/assignees-menu";
import { TableCheckbox } from "@/components/ui";
import { ProjectStatusesMenu } from "./statuses-menu";

export const Project = ({
  name,
  description,
}: {
  name: string;
  description: string;
}) => {
  return (
    <RowWrapper>
      <Flex align="center" className="relative select-none" gap={2}>
        <TableCheckbox />
        <Hexagon className="h-[1.2rem] w-auto" />
        <Link className="flex items-center gap-5" href="/issue">
          <Text className="w-[215px] truncate hover:opacity-90">{name}</Text>
          <Text className="max-w-lg truncate hover:opacity-90" color="muted">
            {description}
          </Text>
        </Link>
      </Flex>
      <Flex align="center" gap={5}>
        <Text className="flex items-center gap-1" fontWeight="medium">
          <ProjectStatusesMenu asIcon={false} status="In Progress" />
        </Text>
        <Tooltip title="Starts on">
          <Text className="flex items-center gap-1" color="muted">
            <Calendar className="h-5 w-auto" strokeWidth={2} />
            Sep 27, 2024
          </Text>
        </Tooltip>
        <Tooltip title="Due on">
          <Text className="flex items-center gap-1" color="muted">
            <CalendarCheck2
              className="h-5 w-auto text-primary"
              strokeWidth={2}
            />
            Sep 27, 2024
          </Text>
        </Tooltip>
        <AssigneesMenu isSearchEnabled placeholder="Assign project lead..." />
      </Flex>
    </RowWrapper>
  );
};
