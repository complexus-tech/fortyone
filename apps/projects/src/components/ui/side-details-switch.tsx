"use client";
import { Button, Tooltip } from "ui";

export const SideDetailsSwitch = ({
  isExpanded,
  setIsExpanded,
  disabled,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
  disabled?: boolean;
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
        onClick={() => {
          setIsExpanded(!isExpanded);
        }}
        size="sm"
        variant="outline"
      >
        Analytics
        <span className="sr-only">
          {isExpanded ? "Hide panel" : "Show panel"}
        </span>
      </Button>
    </Tooltip>
  );
};
