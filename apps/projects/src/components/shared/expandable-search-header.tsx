"use client";

import type { FormEvent, ReactNode } from "react";
import { useState } from "react";
import { SearchIcon } from "icons";
import { Button, Flex, Input } from "ui";

type ExpandableSearchHeaderProps = {
  actions?: ReactNode;
  initialValue: string;
  label: string;
  leading: ReactNode;
  onSubmit: (search: string) => void;
  placeholder: string;
};

export const ExpandableSearchHeader = ({
  actions,
  initialValue,
  label,
  leading,
  onSubmit,
  placeholder,
}: ExpandableSearchHeaderProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [value, setValue] = useState(initialValue);
  const isExpanded = isOpen || value.length > 0;

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const search = value.trim();
    setValue(search);
    onSubmit(search);
  };

  return (
    <Flex
      align="center"
      className="border-border/60 h-16 min-w-0 gap-2 border-b-[0.5px] px-4"
    >
      {!isExpanded ? <div className="min-w-0 flex-1">{leading}</div> : null}
      <form
        className={
          isExpanded
            ? "min-w-0 flex-1 transition-all duration-200 ease-out"
            : "w-[2.1rem] shrink-0 transition-all duration-200 ease-out"
        }
        onSubmit={handleSubmit}
        role="search"
      >
        {isExpanded ? (
          <Input
            aria-label={label}
            autoFocus={isOpen}
            className="h-[2.1rem]"
            leftIcon={<SearchIcon className="h-4" />}
            onBlur={() => {
              setIsOpen(false);
            }}
            onChange={(event) => {
              setValue(event.target.value);
            }}
            onFocus={() => {
              setIsOpen(true);
            }}
            onKeyDown={(event) => {
              if (event.key === "Escape") {
                setValue("");
                setIsOpen(false);
                onSubmit("");
                event.currentTarget.blur();
              }
            }}
            placeholder={placeholder}
            size="sm"
            type="search"
            value={value}
          />
        ) : (
          <Button
            aria-label={label}
            asIcon
            className="aspect-square"
            color="tertiary"
            leftIcon={<SearchIcon className="h-4" />}
            onClick={() => {
              setIsOpen(true);
            }}
            size="sm"
            type="button"
          />
        )}
      </form>
      {actions ? (
        <Flex align="center" className="shrink-0" gap={1}>
          {actions}
        </Flex>
      ) : null}
    </Flex>
  );
};
