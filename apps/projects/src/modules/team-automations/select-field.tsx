import { useId } from "react";
import { Box, Select } from "ui";
import { cn } from "lib";

export const AutomationSelect = ({
  label,
  value,
  onChange,
  options,
  layout = "stacked",
  disabled = false,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
  layout?: "stacked" | "inline" | "compact";
  disabled?: boolean;
}) => {
  const id = useId();
  return (
    <Box
      className={cn({
        "space-y-2": layout === "stacked",
        "min-w-0 space-y-1.5": layout === "compact",
        "grid gap-2 sm:grid-cols-[8rem_minmax(0,1fr)] sm:items-center":
          layout === "inline",
      })}
    >
      <label className="block" htmlFor={id}>
        {label}
      </label>
      <Select disabled={disabled} onValueChange={onChange} value={value}>
        <Select.Trigger
          className={cn("text-base disabled:opacity-50", {
            "h-10 w-full px-3": layout === "stacked",
            "px-2": layout === "compact",
          })}
          id={id}
        >
          <Select.Input />
        </Select.Trigger>
        <Select.Content>
          {options.map((option) => (
            <Select.Option
              className="py-2 text-base"
              key={option.value}
              value={option.value}
            >
              {option.label}
            </Select.Option>
          ))}
        </Select.Content>
      </Select>
    </Box>
  );
};
