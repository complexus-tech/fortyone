"use client";
import { useMemo, useState } from "react";
import { isAfter, isBefore, isThisWeek, isToday } from "date-fns";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useSession } from "next-auth/react";
import { useParams } from "next/navigation";
import { Box, Button, Tabs, Text } from "ui";
import { ArrowUpDownIcon, CopyIcon } from "icons";
import { toast } from "sonner";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { useStatuses } from "@/lib/hooks/statuses";
import { useSprints } from "@/modules/sprints/hooks/sprints";
import { useObjectiveOptions } from "@/modules/objectives/stories/provider";
import { useObjectiveStories } from "@/modules/stories/hooks/objective-stories";
import { useCopyToClipboard } from "@/hooks";
import { Overview } from "./overview";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const session = useSession();
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const [_, copyText] = useCopyToClipboard();
  const { data: stories = [] } = useObjectiveStories(objectiveId);
  const { data: statuses = [] } = useStatuses();
  const { data: sprints = [] } = useSprints();

  const tabs = ["overview", "all", "active", "backlog"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  type Tab = (typeof tabs)[number];
  // a sprint is active if it has a start date and end date and the current date is between the start and end date
  // use date-fns to check if the current date is between the start and end date
  const activeSprints = sprints
    .filter(
      (sprint) =>
        sprint.startDate &&
        sprint.endDate &&
        isAfter(new Date(), new Date(sprint.startDate)) &&
        isBefore(new Date(), new Date(sprint.endDate)),
    )
    .map((sprint) => sprint.id);

  const completedStatuses = statuses
    .filter((state) => state.category === "completed")
    .map((state) => state.id);
  const { viewOptions, filters } = useObjectiveOptions();
  const backlogStatuses = statuses
    .filter((state) => state.category === "backlog")
    .map((state) => state.id);
  const activeStatuses = statuses
    .filter((state) => state.category === "started")
    .map((state) => state.id);

  //filters
  const filteredStories = useMemo(() => {
    let newStories = [...stories];
    if (filters.completed) {
      newStories = newStories.filter((story) =>
        completedStatuses.includes(story.statusId),
      );
    }
    if (filters.dueThisWeek) {
      newStories = newStories.filter((story) =>
        isThisWeek(new Date(story.endDate!)),
      );
    }
    if (filters.dueToday) {
      newStories = newStories.filter((story) =>
        isToday(new Date(story.endDate!)),
      );
    }
    if (filters.completed) {
      newStories = newStories.filter((story) =>
        completedStatuses.includes(story.statusId),
      );
    }
    if (filters.assignedToMe) {
      newStories = newStories.filter(
        (story) => story.assigneeId === session.data?.user?.id,
      );
    }
    if (filters.activeSprints) {
      newStories = newStories.filter((story) =>
        activeSprints.includes(story.sprintId!),
      );
    }
    return newStories;
  }, [stories, filters]);

  const backlog = filteredStories.filter((story) =>
    backlogStatuses.includes(story.statusId),
  );
  const activeStories = filteredStories.filter((story) =>
    activeStatuses.includes(story.statusId),
  );

  const [isCopied, setIsCopied] = useState(false);

  return (
    <Tabs onValueChange={(v) => setTab(v as Tab)} value={tab}>
      <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full items-center justify-between border-b-[0.5px] border-gray-100/60 pr-12 dark:border-dark-100">
        <Tabs.List className="h-min">
          <Tabs.Tab value="overview">Overview</Tabs.Tab>
          <Tabs.Tab value="all">All stories</Tabs.Tab>
          <Tabs.Tab value="active">Active</Tabs.Tab>
          <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
        </Tabs.List>
        {tab !== "overview" ? (
          <Text
            className="ml-2 flex shrink-0 items-center gap-1.5 px-1 opacity-80"
            color="muted"
          >
            <ArrowUpDownIcon className="h-4 w-auto" />
            Ordering by <b className="capitalize">{viewOptions.orderBy}</b>
          </Text>
        ) : (
          <Button
            className="gap-1 px-3"
            color="tertiary"
            leftIcon={<CopyIcon className="h-4" />}
            onClick={async () => {
              await copyText(window.location.href);
              setIsCopied(true);
              toast.info("Success", {
                description: "Objective link copied to clipboard",
              });
              setTimeout(() => {
                setIsCopied(false);
              }, 5000);
            }}
            size="sm"
          >
            {isCopied ? "Copied" : "Copy link"}
          </Button>
        )}
      </Box>
      <Tabs.Panel value="overview">
        <Overview />
      </Tabs.Panel>

      <Tabs.Panel value="all">
        <StoriesBoard
          className="h-[calc(100vh-7.7rem)]"
          layout={layout}
          stories={filteredStories}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
      <Tabs.Panel value="active">
        <StoriesBoard
          className="h-[calc(100vh-7.7rem)]"
          layout={layout}
          stories={activeStories}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
      <Tabs.Panel value="backlog">
        <StoriesBoard
          className="h-[calc(100vh-7.7rem)]"
          layout={layout}
          stories={backlog}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
    </Tabs>
  );
};
