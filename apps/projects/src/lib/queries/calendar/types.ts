export type CalendarBusyWindow = {
  id: string;
  provider: string;
  title?: string;
  startAt: string;
  endAt: string;
  status: "busy";
  isPrivate: boolean;
  createdAt: string;
  updatedAt: string;
};

export type CalendarScheduleBlock = {
  id: string;
  storyId?: string;
  storyTitle?: string;
  storyCode?: string;
  teamId?: string;
  teamName?: string;
  teamCode?: string;
  blockType: "work" | "focus";
  title: string;
  startAt: string;
  endAt: string;
  isLocked: boolean;
  source: "user" | "maya";
  createdAt: string;
  updatedAt: string;
};

export type CalendarSchedule = {
  startAt: string;
  endAt: string;
  busyWindows: CalendarBusyWindow[];
  blocks: CalendarScheduleBlock[];
};

export type CalendarScheduleBlockInput = {
  storyId?: string | null;
  blockType: "work" | "focus";
  title: string;
  startAt: string;
  endAt: string;
  isLocked?: boolean;
};
