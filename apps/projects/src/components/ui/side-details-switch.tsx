"use client";
import { Button, Tooltip } from "ui";
import { SidebarCollapseIcon, SidebarExpandIcon } from "icons";

export const SideDetailsSwitch = ({
  isExpanded,
  setIsExpanded,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
}) => {
  return (
    <Tooltip title={isExpanded ? "Hide sprint details" : "Show sprint details"}>
      <Button
        color="tertiary"
        leftIcon={
          isExpanded ? (
            <SidebarCollapseIcon className="h-5 w-auto" />
          ) : (
            <SidebarExpandIcon className="h-5 w-auto" />
          )
        }
        onClick={() => {
          setIsExpanded(!isExpanded);
        }}
        size="sm"
        variant="naked"
      >
        <span className="sr-only">
          {isExpanded ? "Hide sprint details" : "Show sprint details"}
        </span>
      </Button>
    </Tooltip>
  );
};
