"use client";

import {
  ArrowDown2Icon,
  CalendarIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  SettingsIcon,
  TimeScheduleIcon,
} from "icons";
import { BreadCrumbs, Button, Flex, Menu, Text } from "ui";
import {
  HeaderContainer,
  MobileMenuButton,
  useAppCommandAction,
} from "@/components/shared";
import { useTerminology, useUserRole } from "@/hooks";
import type { CalendarView } from "./calendar-view";

const calendarViews = ["day", "week", "month"] as const;
const calendarViewLabels = {
  day: "Day",
  month: "Month",
  week: "Week",
} satisfies Record<CalendarView, string>;

type CalendarHeaderProps = {
  canNavigateNext: boolean;
  canNavigatePrevious: boolean;
  currentView: CalendarView;
  onFocus: () => void;
  onNext: () => void;
  onPrevious: () => void;
  onSchedule: () => void;
  onShowCompletedWorkChange: (show: boolean) => void;
  onToday: () => void;
  onViewChange: (view: CalendarView) => void;
  showCompletedWork: boolean;
  title: string;
};

export const CalendarHeader = ({
  canNavigateNext,
  canNavigatePrevious,
  currentView,
  onFocus,
  onNext,
  onPrevious,
  onSchedule,
  onShowCompletedWorkChange,
  onToday,
  onViewChange,
  showCompletedWork,
  title,
}: CalendarHeaderProps) => {
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();

  useAppCommandAction({
    disabled: userRole === "guest",
    id: "calendar:schedule-story",
    label: `Schedule ${getTermDisplay("storyTerm")}`,
    onSelect: onSchedule,
  });

  return (
    <HeaderContainer className="shrink-0 justify-between gap-4 overflow-x-auto">
      <Flex align="center" className="shrink-0" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Calendar",
              icon: <CalendarIcon />,
            },
            {
              name: calendarViewLabels[currentView],
            },
          ]}
        />
      </Flex>
      <Flex align="center" className="shrink-0" gap={2}>
        <Flex align="center" gap={1}>
          <Button
            aria-label={`Previous ${currentView}`}
            asIcon
            className="focus-visible:ring-primary/40 focus-visible:ring-2"
            color="tertiary"
            disabled={!canNavigatePrevious}
            onClick={onPrevious}
            size="sm"
            variant="naked"
          >
            <ChevronLeftIcon className="h-5" />
          </Button>
          <Button
            aria-label={`Next ${currentView}`}
            asIcon
            className="focus-visible:ring-primary/40 focus-visible:ring-2"
            color="tertiary"
            disabled={!canNavigateNext}
            onClick={onNext}
            size="sm"
            variant="naked"
          >
            <ChevronRightIcon className="h-5" />
          </Button>
        </Flex>
        <Text className="mr-1 whitespace-nowrap" fontWeight="medium">
          {title}
        </Text>
        <Button color="tertiary" onClick={onToday} size="sm">
          Today
        </Button>
        <Menu>
          <Menu.Button>
            <Button
              className="justify-between capitalize"
              color="tertiary"
              rightIcon={<ArrowDown2Icon className="h-4" />}
              size="sm"
              variant="outline"
            >
              {currentView}
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-36">
            <Menu.Group>
              {calendarViews.map((view) => (
                <Menu.Item
                  active={currentView === view}
                  className="py-2.5 text-base capitalize"
                  key={view}
                  onSelect={() => {
                    onViewChange(view);
                  }}
                >
                  {view}
                </Menu.Item>
              ))}
            </Menu.Group>
          </Menu.Items>
        </Menu>
        <Menu>
          <Menu.Button>
            <Button
              aria-label="Calendar display options"
              asIcon
              color="tertiary"
              size="sm"
              variant="outline"
            >
              <SettingsIcon className="h-[1.1rem]" />
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-56">
            <Menu.Group>
              <Menu.CheckboxItem
                checked={showCompletedWork}
                onCheckedChange={(checked) => {
                  onShowCompletedWorkChange(checked);
                }}
              >
                <span className="pl-6">Show completed work</span>
              </Menu.CheckboxItem>
            </Menu.Group>
          </Menu.Items>
        </Menu>
        <span className="text-text-secondary mx-1 hidden opacity-40 md:inline">
          |
        </span>
        <Button
          color="tertiary"
          leftIcon={<TimeScheduleIcon className="h-[1.1rem]" strokeWidth={2} />}
          onClick={onFocus}
          size="sm"
          variant="outline"
        >
          Block focus time
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
