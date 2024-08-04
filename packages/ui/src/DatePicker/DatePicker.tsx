"use client";
import { ReactNode, createContext, useContext, useState } from "react";
import { Calendar, CalendarProps } from "../Calendar/Calendar";
import { Popover } from "../Popover/Popover";

const PickerContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const usePicker = () => {
  const { open, setOpen } = useContext(PickerContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = usePicker();
  return (
    <Popover open={open} onOpenChange={setOpen}>
      {children}
    </Popover>
  );
};

export const DatePicker = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <PickerContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </PickerContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => {
  return <Popover.Trigger asChild>{children}</Popover.Trigger>;
};

const CalendarPrimitive = (props: CalendarProps) => {
  const { onDayClick, ...rest } = props;
  const { setOpen } = usePicker();
  return (
    <Popover.Content className="w-auto p-0 z-50 rounded-xl border-0 bg-transparent dark:bg-transparent">
      <Calendar
        {...rest}
        onDayClick={(day, activeModifiers, event) => {
          if (!onDayClick) return;
          onDayClick(day, activeModifiers, event);
          setOpen(false);
        }}
      />
    </Popover.Content>
  );
};

DatePicker.Trigger = Trigger;
DatePicker.Calendar = CalendarPrimitive;
