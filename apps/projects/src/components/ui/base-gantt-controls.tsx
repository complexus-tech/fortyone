"use client";

import { ArrowDown2Icon } from "icons";
import type { ReactNode } from "react";
import { Box, Button, Flex, Menu, Text } from "ui";
import type { ZoomLevel } from "./base-gantt-utils";

const ZOOM_LABELS: Record<ZoomLevel, string> = {
  months: "Months",
  quarters: "Quarters",
  weeks: "Weeks",
};
const ZOOM_LEVELS: ZoomLevel[] = ["weeks", "months", "quarters"];

export const GanttHeader = ({
  onReset,
  zoomLevel,
  onZoomChange,
  children,
}: {
  onReset: () => void;
  zoomLevel: ZoomLevel;
  onZoomChange: (zoom: ZoomLevel) => void;
  children?: ReactNode;
}) => (
  <Box className="border-border bg-background sticky top-0 z-10 hidden h-16 border-b-[0.5px] px-4 md:block">
    <GanttControls
      className={children ? "h-9" : "h-full"}
      onReset={onReset}
      onZoomChange={onZoomChange}
      zoomLevel={zoomLevel}
    />
    {children}
  </Box>
);

export const GanttControls = ({
  onReset,
  zoomLevel,
  onZoomChange,
  className,
  showSeparator = false,
}: {
  onReset: () => void;
  zoomLevel: ZoomLevel;
  onZoomChange: (zoom: ZoomLevel) => void;
  className?: string;
  showSeparator?: boolean;
}) => (
  <Flex align="center" className={className} gap={2}>
    <Flex align="center" gap={2}>
      <Text color="muted" fontWeight="medium">
        Zoom:
      </Text>
      <Menu>
        <Menu.Button>
          <Button
            color="tertiary"
            rightIcon={<ArrowDown2Icon className="h-4" />}
            size="sm"
          >
            {ZOOM_LABELS[zoomLevel]}
          </Button>
        </Menu.Button>
        <Menu.Items className="w-40">
          <Menu.Group>
            {ZOOM_LEVELS.map((zoom) => (
              <Menu.Item
                key={zoom}
                onSelect={() => {
                  onZoomChange(zoom);
                }}
              >
                {ZOOM_LABELS[zoom]}
              </Menu.Item>
            ))}
          </Menu.Group>
        </Menu.Items>
      </Menu>
    </Flex>
    {showSeparator ? (
      <span className="text-text-secondary mx-1 hidden opacity-40 md:inline">
        |
      </span>
    ) : null}
    <Button color="tertiary" onClick={onReset} size="sm">
      Today
    </Button>
  </Flex>
);
