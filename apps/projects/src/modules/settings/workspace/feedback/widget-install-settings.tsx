"use client";

import { useMemo, useState, useSyncExternalStore } from "react";
import { CheckIcon, CopyIcon } from "icons";
import { Box, Button, Flex, Input, Text } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import {
  buildFeedbackWidgetSnippet,
  buildInlineFeedbackWidgetMarkup,
} from "@/modules/feedback-widget/install";
import type {
  FeedbackWidgetMode,
  FeedbackWidgetTab,
  FeedbackWidgetTheme,
} from "@/modules/feedback-widget/protocol";

const selectClassName =
  "border-border bg-surface-elevated ring-ring h-10 w-full rounded-xl border px-3 text-sm outline-none focus-visible:ring-2";

const subscribeToBrowserOrigin = () => () => undefined;
const getBrowserOrigin = () => window.location.origin;
const getServerOrigin = () => "";

export const WidgetInstallSettings = ({
  portalSlug,
  scriptOrigin,
}: {
  portalSlug: string;
  scriptOrigin?: string;
}) => {
  const [mode, setMode] = useState<FeedbackWidgetMode>("bubble");
  const [defaultTab, setDefaultTab] = useState<FeedbackWidgetTab>("feedback");
  const [theme, setTheme] = useState<FeedbackWidgetTheme>("auto");
  const [position, setPosition] = useState<"bottom-left" | "bottom-right">(
    "bottom-right",
  );
  const [trigger, setTrigger] = useState("[data-fortyone-feedback]");
  const [copied, setCopied] = useState(false);
  const browserOrigin = useSyncExternalStore(
    subscribeToBrowserOrigin,
    getBrowserOrigin,
    getServerOrigin,
  );
  const resolvedScriptOrigin = scriptOrigin || browserOrigin || undefined;
  const snippet = useMemo(() => {
    const options = {
      defaultTab,
      portalSlug,
      position,
      scriptOrigin: resolvedScriptOrigin,
      theme,
      trigger,
    };
    return mode === "inline"
      ? buildInlineFeedbackWidgetMarkup(options)
      : buildFeedbackWidgetSnippet({ ...options, mode });
  }, [
    defaultTab,
    mode,
    portalSlug,
    position,
    resolvedScriptOrigin,
    theme,
    trigger,
  ]);

  const copySnippet = async () => {
    try {
      await navigator.clipboard.writeText(snippet);
      setCopied(true);
      window.setTimeout(() => {
        setCopied(false);
      }, 1800);
    } catch {
      setCopied(false);
    }
  };

  return (
    <Box className="border-border bg-surface mb-6 overflow-hidden rounded-2xl border">
      <SectionHeader
        description="Add Feedback and Roadmap to your product without sending customers to a separate page."
        title="Feedback widget"
      />
      <Box className="space-y-6 p-6">
        <Box className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <label className="space-y-2">
            <Text as="span" className="block" fontWeight="medium">
              Display
            </Text>
            <select
              aria-label="Widget display"
              className={selectClassName}
              onChange={(event) => {
                setMode(event.target.value as FeedbackWidgetMode);
              }}
              value={mode}
            >
              <option value="bubble">Floating bubble</option>
              <option value="custom">Custom trigger</option>
              <option value="inline">Inline</option>
            </select>
          </label>
          <label className="space-y-2">
            <Text as="span" className="block" fontWeight="medium">
              Opens on
            </Text>
            <select
              aria-label="Widget default tab"
              className={selectClassName}
              onChange={(event) => {
                setDefaultTab(event.target.value as FeedbackWidgetTab);
              }}
              value={defaultTab}
            >
              <option value="feedback">Feedback</option>
              <option value="roadmap">Roadmap</option>
            </select>
          </label>
          <label className="space-y-2">
            <Text as="span" className="block" fontWeight="medium">
              Theme
            </Text>
            <select
              aria-label="Widget theme"
              className={selectClassName}
              onChange={(event) => {
                setTheme(event.target.value as FeedbackWidgetTheme);
              }}
              value={theme}
            >
              <option value="auto">Match device</option>
              <option value="light">Light</option>
              <option value="dark">Dark</option>
            </select>
          </label>
          <label className="space-y-2">
            <Text as="span" className="block" fontWeight="medium">
              Position
            </Text>
            <select
              aria-label="Widget position"
              className={selectClassName}
              disabled={mode === "inline"}
              onChange={(event) => {
                setPosition(event.target.value as typeof position);
              }}
              value={position}
            >
              <option value="bottom-right">Bottom right</option>
              <option value="bottom-left">Bottom left</option>
            </select>
          </label>
        </Box>

        {mode === "custom" ? (
          <Box className="max-w-lg space-y-2">
            <label
              className="text-foreground block font-medium"
              htmlFor="feedback-widget-trigger"
            >
              Trigger selector
            </label>
            <Input
              id="feedback-widget-trigger"
              onChange={(event) => {
                setTrigger(event.target.value);
              }}
              value={trigger}
            />
            <Text color="muted">
              Clicking any matching element opens the widget. The loader waits
              for triggers added after page load too.
            </Text>
          </Box>
        ) : null}

        <Box className="border-border/70 bg-background overflow-hidden rounded-xl border">
          <Flex
            align="center"
            className="border-border/70 border-b px-4 py-3"
            justify="between"
          >
            <Text fontWeight="medium">Installation snippet</Text>
            <Button
              color="tertiary"
              leftIcon={
                copied ? (
                  <CheckIcon className="h-4" />
                ) : (
                  <CopyIcon className="h-4" />
                )
              }
              onClick={() => {
                void copySnippet();
              }}
              size="sm"
            >
              {copied ? "Copied" : "Copy code"}
            </Button>
          </Flex>
          <pre className="max-h-72 overflow-auto p-4 text-[12px] leading-6 font-normal whitespace-pre-wrap">
            <code>{snippet}</code>
          </pre>
        </Box>
        <Text color="muted">
          Paste this before the closing body tag. The script loads
          asynchronously and does not download the widget interface until
          someone opens it.
        </Text>
      </Box>
    </Box>
  );
};
