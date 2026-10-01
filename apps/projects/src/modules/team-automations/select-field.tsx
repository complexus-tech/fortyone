import { useId } from "react";
import { Box, Select } from "ui";

export const AutomationSelect = ({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
}) => {
  const id = useId();
  return (
    <Box className="space-y-2">
      <label className="block" htmlFor={id}>
        {label}
      </label>
      <Select onValueChange={onChange} value={value}>
        <Select.Trigger className="h-10 w-full px-3 text-base" id={id}>
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
