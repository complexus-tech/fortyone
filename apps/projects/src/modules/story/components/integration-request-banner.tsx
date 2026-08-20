"use client";

import { useState } from "react";
import Link from "next/link";
import { CopyIcon, LinkIcon, MoreHorizontalIcon, SlackIcon } from "icons";
import { cn } from "lib";
import { Box, Button, Collapsible, Flex, Menu, Text } from "ui";
import { useWorkspacePath } from "@/hooks";
import { IntegrationRequestThreadActivity } from "@/modules/integration-requests/thread-activity";
import type { IntegrationRequestProviderThread } from "@/modules/integration-requests/types";

export const IntegrationRequestBannerRow = ({
  embedded = false,
  link,
}: {
  embedded?: boolean;
  link: IntegrationRequestProviderThread;
}) => {
  const [isThreadOpen, setIsThreadOpen] = useState(true);
  const { withWorkspace } = useWorkspacePath();
  const requestHref = withWorkspace(
    `/teams/${link.teamId}/requests/${link.integrationRequestId}`,
  );

  return (
    <Collapsible onOpenChange={setIsThreadOpen} open={isThreadOpen}>
      <Flex
        align="center"
        className={cn(
          "border-primary/20 bg-primary/5 rounded-xl border px-4 py-3",
          {
            "rounded-b-none": isThreadOpen && !embedded,
            "rounded-none border-0 bg-transparent": embedded,
          },
        )}
        justify="between"
      >
        <Collapsible.Trigger asChild>
          <Button
            aria-label={`Toggle Slack conversation for ${link.requestTitle}`}
            className="text-primary -my-3 h-auto min-w-0 flex-1 justify-start gap-2 rounded-lg px-0 py-3 text-left hover:bg-transparent"
            color="tertiary"
            leftIcon={<SlackIcon className="h-5 shrink-0" />}
            size="sm"
            variant="naked"
          >
            <Text
              as="span"
              className="min-w-0 truncate"
              color="primary"
              fontWeight="medium"
            >
              From Slack request · {link.requestTitle}
            </Text>
          </Button>
        </Collapsible.Trigger>
        <Flex align="center" className="shrink-0" gap={1}>
          <Link
            aria-label="Open Slack request"
            className="text-primary hover:text-primary/80 rounded-md p-1 transition"
            href={requestHref}
          >
            <LinkIcon className="text-current" />
          </Link>
          <Menu>
            <Menu.Button>
              <button
                aria-label="More Slack request actions"
                className="text-primary hover:text-primary/80 rounded-md p-1 transition"
                type="button"
              >
                <MoreHorizontalIcon className="h-5 text-current" />
              </button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item
                  onSelect={() => {
                    navigator.clipboard.writeText(
                      new URL(requestHref, window.location.origin).toString(),
                    );
                  }}
                >
                  <CopyIcon />
                  Copy request link
                </Menu.Item>
                {link.sourceUrl ? (
                  <Menu.Item
                    onSelect={() => {
                      window.open(
                        link.sourceUrl,
                        "_blank",
                        "noopener,noreferrer",
                      );
                    }}
                  >
                    <SlackIcon className="h-5" />
                    Open in Slack
                  </Menu.Item>
                ) : null}
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      </Flex>
      <Collapsible.Content>
        <Box
          className={cn(
            "border-primary/20 bg-surface/50 rounded-b-xl border border-t-0 px-4 py-4",
            { "rounded-none border-x-0 border-b-0": embedded },
          )}
        >
          <IntegrationRequestThreadActivity
            compact
            requestId={link.integrationRequestId}
          />
        </Box>
      </Collapsible.Content>
    </Collapsible>
  );
};

export const IntegrationRequestBanner = ({
  links,
}: {
  links: IntegrationRequestProviderThread[];
}) => (
  <Box className="mb-3 space-y-2">
    {links.map((link) => (
      <IntegrationRequestBannerRow key={link.id} link={link} />
    ))}
  </Box>
);
