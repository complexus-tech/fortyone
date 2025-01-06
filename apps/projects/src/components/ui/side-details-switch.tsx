"use client";
import { Button, Tooltip } from "ui";

export const SideDetailsSwitch = ({
  isExpanded,
  setIsExpanded,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
}) => {
  return (
    <Tooltip title={isExpanded ? "Hide panel" : "Show panel"}>
      <Button
        color="tertiary"
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
