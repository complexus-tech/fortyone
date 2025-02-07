"use client";

import { cn } from "lib";
import { Box, Text } from "ui";

const DOMAINS = [
  { id: "engineering", label: "Engineering" },
  { id: "product", label: "Product" },
  { id: "marketing", label: "Marketing" },
  { id: "sales", label: "Sales" },
  { id: "legal", label: "Legal" },
  { id: "finance", label: "Finance" },
  { id: "hr", label: "Human Resources" },
  { id: "other", label: "Other" },
] as const;

type DomainSelectorProps = {
  value: string;
  onChange: (value: string) => void;
};

export const DomainSelector = ({ value, onChange }: DomainSelectorProps) => {
  return (
    <Box className="mt-8">
      <Box className="mt-3 grid grid-cols-2 gap-3">
        {DOMAINS.map((domain) => (
          <label
            className="flex cursor-pointer items-center rounded-xl border border-gray-100 p-3.5 ring-primary transition duration-200 ease-linear hover:ring-[2px] dark:border-dark-200 dark:ring-offset-dark"
            key={domain.id}
          >
            <input
              checked={value === domain.id}
              className="sr-only"
              name="domain"
              onChange={() => {
                onChange(domain.id);
              }}
              type="radio"
              value={domain.id}
            />
            <Text className="flex-1 font-medium">{domain.label}</Text>

            <Box
              className={cn(
                "ml-3 size-4 rounded-full border-2 border-gray-300 dark:border-dark-100",
                {
                  "border-primary dark:border-primary": value === domain.id,
                },
              )}
            />
          </label>
        ))}
      </Box>
    </Box>
  );
};
