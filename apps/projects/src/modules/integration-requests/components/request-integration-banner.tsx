"use client";

import type { ReactNode } from "react";
import {
  CheckIcon,
  CloseIcon,
  CopyIcon,
  LinkIcon,
  MoreHorizontalIcon,
} from "icons";
import { Box, Flex, Menu, Text } from "ui";

export type RequestSourceBannerDetails = {
  icon: ReactNode;
  openLabel: string;
  primaryText: string;
  secondaryText: string | null;
};

export const RequestIntegrationBanner = ({
  canEditRequest,
  icon,
  onAccept,
  onDecline,
  openLabel,
  primaryText,
  secondaryText,
  sourceUrl,
}: {
  canEditRequest: boolean;
  icon: ReactNode;
  onAccept: () => void;
  onDecline: () => void;
  openLabel: string;
  primaryText: string;
  secondaryText: string | null;
  sourceUrl?: string;
}) => (
  <Box className="mb-3 space-y-2">
    <Flex
      align="center"
      className="border-primary/20 bg-primary/5 rounded-xl border px-4 py-3"
      justify="between"
    >
      <Flex align="center" className="min-w-0" gap={2}>
        {icon}
        <Text className="line-clamp-1" color="primary" fontWeight="medium">
          {primaryText}
        </Text>
        {secondaryText ? (
          <Text className="line-clamp-1" color="muted">
            {secondaryText}
          </Text>
        ) : null}
      </Flex>
      <Flex align="center" className="shrink-0" gap={1}>
        {sourceUrl ? (
          <a
            className="text-primary hover:text-primary/80 rounded-md p-1 transition"
            href={sourceUrl}
            rel="noopener noreferrer"
            target="_blank"
            title={openLabel}
          >
            <LinkIcon className="text-current" />
          </a>
        ) : null}
        <Menu>
          <Menu.Button>
            <button
              aria-label="Intake actions"
              className="text-primary hover:text-primary/80 rounded-md p-1 transition"
              type="button"
            >
              <MoreHorizontalIcon className="h-5 text-current" />
            </button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group>
              {sourceUrl ? (
                <Menu.Item
                  onSelect={() => {
                    window.open(sourceUrl, "_blank", "noopener,noreferrer");
                  }}
                >
                  <LinkIcon className="text-icon h-5 w-auto" />
                  {openLabel}
                </Menu.Item>
              ) : null}
              {sourceUrl ? (
                <Menu.Item
                  onSelect={() => {
                    navigator.clipboard.writeText(sourceUrl);
                  }}
                >
                  <CopyIcon className="text-icon h-5 w-auto" />
                  Copy link
                </Menu.Item>
              ) : null}
              <Menu.Item disabled={!canEditRequest} onSelect={onAccept}>
                <CheckIcon className="text-icon h-5 w-auto" />
                Accept intake item
              </Menu.Item>
              <Menu.Item
                className="text-danger"
                disabled={!canEditRequest}
                onSelect={onDecline}
              >
                <CloseIcon className="text-danger" />
                Decline intake item...
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Flex>
  </Box>
);
