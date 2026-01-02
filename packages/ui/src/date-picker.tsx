"use client";
import { ReactNode, createContext, useContext, useState } from "react";
import { Calendar, CalendarProps } from "./calendar";
import { Popover } from "./popover";

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
    <Popover.Content className="w-auto p-0 rounded-2xl border border-border bg-surface-elevated z-50 backdrop-blur">
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
