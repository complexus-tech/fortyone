import { subDays, startOfDay, endOfDay } from "date-fns";

export type DatePreset = {
  label: string;
  value: string;
  getDates: () => { startDate: Date; endDate: Date };
};

export const datePresets: DatePreset[] = [
  {
    label: "Last 7 days",
    value: "7d",
    getDates: () => ({
      startDate: startOfDay(subDays(new Date(), 7)),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: "Last 30 days",
    value: "30d",
    getDates: () => ({
      startDate: startOfDay(subDays(new Date(), 30)),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: "Last 90 days",
    value: "90d",
    getDates: () => ({
      startDate: startOfDay(subDays(new Date(), 90)),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: "This month",
    value: "month",
    getDates: () => {
      const now = new Date();
      const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
      return {
        startDate: startOfDay(startOfMonth),
        endDate: endOfDay(now),
      };
    },
  },
];

export type FilterButtonProps = {
  label: string;
  icon: React.ReactNode;
  text: string;
  popover: React.ReactNode;
  isActive?: boolean;
};
