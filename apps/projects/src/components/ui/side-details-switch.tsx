"use client";
import { DashboardIcon } from "icons";
import { Button, Tooltip } from "ui";

export const SideDetailsSwitch = ({
  isExpanded,
  setIsExpanded,
  disabled,
  label = "Insights",
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
  disabled?: boolean;
  label?: string;
}) => {
  const title = () => {
    if (disabled) return null;
    if (isExpanded) return "Hide panel";
    return "Show panel";
  };

  return (
    <Tooltip title={title()}>
      <Button
        color="tertiary"
        disabled={disabled}
        leftIcon={<DashboardIcon className="h-[1.1rem]" />}
        onClick={() => {
          setIsExpanded(!isExpanded);
        }}
        size="sm"
        variant="outline"
      >
        {label}
        <span className="sr-only">
          {isExpanded ? "Hide panel" : "Show panel"}
        </span>
      </Button>
    </Tooltip>
  );
};
