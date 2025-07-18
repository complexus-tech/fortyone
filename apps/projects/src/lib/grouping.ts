import {
  isToday,
  isYesterday,
  isThisWeek,
  isThisMonth,
  isThisYear,
  format,
  parseISO,
} from "date-fns";

export interface GroupableItem {
  createdAt: string | Date;
  [key: string]: unknown;
}

export interface ItemGroup<T extends GroupableItem> {
  label: string;
  items: T[];
  sortOrder: number;
}

// Simple grouping strategy - just what you need
const createGroupingStrategy = () => [
  {
    name: "today",
    condition: (date: Date) => isToday(date),
    getLabel: () => "Today",
    getSortOrder: () => 0,
  },
  {
    name: "yesterday",
    condition: (date: Date) => isYesterday(date),
    getLabel: () => "Yesterday",
    getSortOrder: () => 1,
  },
  {
    name: "thisWeek",
    condition: (date: Date) =>
      isThisWeek(date) && !isToday(date) && !isYesterday(date),
    getLabel: () => "This Week",
    getSortOrder: () => 2,
  },
  {
    name: "thisMonth",
    condition: (date: Date) => isThisMonth(date) && !isThisWeek(date),
    getLabel: () => "This Month",
    getSortOrder: () => 3,
  },
  {
    name: "thisYear",
    condition: (date: Date) => isThisYear(date) && !isThisMonth(date),
    getLabel: (date: Date) => format(date, "MMMM"),
    getSortOrder: (date: Date) => 4 + (11 - date.getMonth()),
  },
  {
    name: "older",
    condition: (date: Date) => !isThisYear(date),
    getLabel: (date: Date) => format(date, "MMMM yyyy"),
    getSortOrder: (date: Date) =>
      20 +
      (new Date().getFullYear() - date.getFullYear()) * 12 +
      (11 - date.getMonth()),
  },
];

const parseDate = (date: string | Date): Date => {
  return typeof date === "string" ? parseISO(date) : date;
};

export function groupChatsByDate<T extends GroupableItem>(
  chats: T[],
): ItemGroup<T>[] {
  const strategies = createGroupingStrategy();
  const groups: Record<string, ItemGroup<T>> = {};

  // Group chats
  chats.forEach((chat) => {
    const chatDate = parseDate(chat.createdAt);

    // Find matching strategy
    const strategy = strategies.find((s) => s.condition(chatDate));

    if (!strategy) {
      const fallbackKey = format(chatDate, "MMMM yyyy");
      groups[fallbackKey].items.push(chat);
      return;
    }

    const groupKey = strategy.getLabel(chatDate);

    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition -- will come back to this
    if (!groups[groupKey]) {
      groups[groupKey] = {
        label: groupKey,
        items: [],
        sortOrder: strategy.getSortOrder(chatDate),
      };
    }

    groups[groupKey].items.push(chat);
  });

  // Sort items within each group (newest first)
  Object.values(groups).forEach((group) => {
    group.items.sort((a, b) => {
      const dateA = parseDate(a.createdAt);
      const dateB = parseDate(b.createdAt);
      return dateB.getTime() - dateA.getTime();
    });
  });

  // Return sorted groups
  return Object.values(groups).sort((a, b) => a.sortOrder - b.sortOrder);
}
