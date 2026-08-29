"use client";

import type { ReactNode } from "react";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { Box, Flex, Input, Select, Text } from "ui";
import { cn } from "lib";
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

const subscribeToBrowserOrigin = () => () => undefined;
const getBrowserOrigin = () => window.location.origin;
const getServerOrigin = () => "";

type CodeTokenKind =
  | "attribute"
  | "comment"
  | "keyword"
  | "number"
  | "plain"
  | "punctuation"
  | "string"
  | "tag";

type CodeToken = {
  kind: CodeTokenKind;
  value: string;
};

const CODE_TOKEN_PATTERN =
  /(?:<!--[\s\S]*?-->|\/\*[\s\S]*?\*\/|\/\/[^\n]*|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|`(?:\\.|[^`\\])*`|\b(?:async|await|const|else|export|false|from|function|if|import|let|new|null|return|true|undefined|var)\b|\b\d+(?:\.\d+)?\b|<\/?[A-Za-z][\w-]*|\b[A-Za-z_:][\w:.-]*(?=\s*=)|[{}()[\].,;:+*=/<>-])/g;

const codeTokenClassNames: Record<CodeTokenKind, string> = {
  attribute: "text-text-secondary",
  comment: "text-text-muted",
  keyword: "text-foreground font-medium",
  number: "text-primary",
  plain: "text-foreground",
  punctuation: "text-text-muted",
  string: "text-primary",
  tag: "text-foreground font-medium",
};

const getCodeTokenKind = (value: string): CodeTokenKind => {
  if (
    value.startsWith("//") ||
    value.startsWith("/*") ||
    value.startsWith("<!--")
  ) {
    return "comment";
  }
  if (['"', "'", "`"].includes(value[0])) return "string";
  if (/^\d/.test(value)) return "number";
  if (value.startsWith("<")) return "tag";
  if (
    /^[A-Za-z_:]/.test(value) &&
    !/^\b(?:async|await|const|else|export|false|from|function|if|import|let|new|null|return|true|undefined|var)\b$/.test(
      value,
    )
  ) {
    return "attribute";
  }
  if (/^[A-Za-z]/.test(value)) return "keyword";
  return "punctuation";
};

const tokenizeCode = (value: string): CodeToken[] => {
  const tokens: CodeToken[] = [];
  let lastIndex = 0;

  for (const match of value.matchAll(CODE_TOKEN_PATTERN)) {
    const { index } = match;
    if (index > lastIndex) {
      tokens.push({ kind: "plain", value: value.slice(lastIndex, index) });
    }
    tokens.push({ kind: getCodeTokenKind(match[0]), value: match[0] });
    lastIndex = index + match[0].length;
  }

  if (lastIndex < value.length) {
    tokens.push({ kind: "plain", value: value.slice(lastIndex) });
  }

  return tokens;
};

const WidgetCodeBlock = ({
  code,
  language,
  maxHeight = "max-h-72",
}: {
  code: string;
  language: string;
  maxHeight?: string;
}) => {
  const [copyLabel, setCopyLabel] = useState("Copy");
  const resetTimer = useRef<number | null>(null);
  const tokens = tokenizeCode(code);

  useEffect(
    () => () => {
      if (resetTimer.current) window.clearTimeout(resetTimer.current);
    },
    [],
  );

  const handleCopy = async () => {
    if (resetTimer.current) window.clearTimeout(resetTimer.current);
    try {
      await navigator.clipboard.writeText(code);
      setCopyLabel("Copied");
    } catch {
      setCopyLabel("Couldn’t copy");
    }
    resetTimer.current = window.setTimeout(() => {
      setCopyLabel("Copy");
      resetTimer.current = null;
    }, 1600);
  };

  return (
    <figure className="bg-surface-muted/70 text-foreground overflow-hidden rounded-xl dark:bg-white/[0.04]">
      <figcaption className="bg-surface-elevated/60 text-text-muted flex min-h-10 items-center justify-between px-4 font-mono text-sm">
        <span>{language}</span>
        <button
          aria-live="polite"
          className="hover:text-foreground focus-visible:ring-ring rounded-md px-2 py-1.5 text-inherit transition-colors focus-visible:ring-1 focus-visible:outline-none"
          onClick={() => {
            void handleCopy();
          }}
          type="button"
        >
          {copyLabel}
        </button>
      </figcaption>
      <pre
        className={cn(
          "overflow-auto px-5 py-5 font-mono text-base leading-7 whitespace-pre",
          maxHeight,
        )}
      >
        <code>
          {tokens.map((token, index) => (
            <span className={codeTokenClassNames[token.kind]} key={index}>
              {token.value}
            </span>
          ))}
        </code>
      </pre>
    </figure>
  );
};

const WidgetOptionRow = ({
  children,
  description,
  label,
}: {
  children: ReactNode;
  description: string;
  label: string;
}) => (
  <Flex
    align="center"
    className="flex-col items-stretch gap-4 px-6 py-4 sm:flex-row sm:items-center"
    justify="between"
  >
    <Box className="min-w-0 flex-1">
      <Text>{label}</Text>
      <Text className="max-w-xl" color="muted" fontSize="sm">
        {description}
      </Text>
    </Box>
    <Box className="shrink-0">{children}</Box>
  </Flex>
);

export const WidgetInstallSettings = ({
  portalSlug,
  scriptOrigin,
}: {
  portalSlug: string;
  scriptOrigin?: string;
}) => {
  const [mode, setMode] = useState<FeedbackWidgetMode>("bubble");
  const [defaultTab, setDefaultTab] = useState<FeedbackWidgetTab>("home");
  const [theme, setTheme] = useState<FeedbackWidgetTheme>("auto");
  const [position, setPosition] = useState<"bottom-left" | "bottom-right">(
    "bottom-right",
  );
  const [trigger, setTrigger] = useState("[data-fortyone-feedback]");
  const browserOrigin = useSyncExternalStore(
    subscribeToBrowserOrigin,
    getBrowserOrigin,
    getServerOrigin,
  );
  const resolvedScriptOrigin = scriptOrigin || browserOrigin || undefined;
  const options = {
    defaultTab,
    portalSlug,
    position,
    scriptOrigin: resolvedScriptOrigin,
    theme,
    trigger,
  };
  const snippet =
    mode === "inline"
      ? buildInlineFeedbackWidgetMarkup(options)
      : buildFeedbackWidgetSnippet({ ...options, mode });

  const selectTriggerClassName = "w-max text-[0.9rem] md:text-base";

  return (
    <>
      <Box className="border-border bg-surface overflow-hidden rounded-2xl border">
        <SectionHeader
          description="Choose how the widget opens and what customers see first."
          title="Appearance"
        />
        <Box className="divide-border divide-y-[0.5px]">
          <WidgetOptionRow
            description="Use a floating launcher, your own button, or place the widget directly in a page."
            label="Display"
          >
            <Select
              onValueChange={(value) => {
                setMode(value as FeedbackWidgetMode);
              }}
              value={mode}
            >
              <Select.Trigger
                aria-label="Widget display"
                className={selectTriggerClassName}
              >
                <Select.Input />
              </Select.Trigger>
              <Select.Content align="end">
                <Select.Option className="text-base" value="bubble">
                  Floating bubble
                </Select.Option>
                <Select.Option className="text-base" value="custom">
                  Custom trigger
                </Select.Option>
                <Select.Option className="text-base" value="inline">
                  Inline
                </Select.Option>
              </Select.Content>
            </Select>
          </WidgetOptionRow>
          <WidgetOptionRow
            description="Choose the section customers land on when the widget opens."
            label="Opens on"
          >
            <Select
              onValueChange={(value) => {
                setDefaultTab(value as FeedbackWidgetTab);
              }}
              value={defaultTab}
            >
              <Select.Trigger
                aria-label="Widget default tab"
                className={selectTriggerClassName}
              >
                <Select.Input />
              </Select.Trigger>
              <Select.Content align="end">
                <Select.Option className="text-base" value="home">
                  Home
                </Select.Option>
                <Select.Option className="text-base" value="feedback">
                  Feedback
                </Select.Option>
                <Select.Option className="text-base" value="roadmap">
                  Roadmap
                </Select.Option>
                <Select.Option className="text-base" value="updates">
                  Updates
                </Select.Option>
              </Select.Content>
            </Select>
          </WidgetOptionRow>
          <WidgetOptionRow
            description="Follow the visitor’s device or force a light or dark appearance."
            label="Theme"
          >
            <Select
              onValueChange={(value) => {
                setTheme(value as FeedbackWidgetTheme);
              }}
              value={theme}
            >
              <Select.Trigger
                aria-label="Widget theme"
                className={selectTriggerClassName}
              >
                <Select.Input />
              </Select.Trigger>
              <Select.Content align="end">
                <Select.Option className="text-base" value="auto">
                  Match device
                </Select.Option>
                <Select.Option className="text-base" value="light">
                  Light
                </Select.Option>
                <Select.Option className="text-base" value="dark">
                  Dark
                </Select.Option>
              </Select.Content>
            </Select>
          </WidgetOptionRow>
          {mode !== "inline" ? (
            <WidgetOptionRow
              description="Place the floating launcher on the left or right side of the screen."
              label="Position"
            >
              <Select
                onValueChange={(value) => {
                  setPosition(value as typeof position);
                }}
                value={position}
              >
                <Select.Trigger
                  aria-label="Widget position"
                  className={selectTriggerClassName}
                >
                  <Select.Input />
                </Select.Trigger>
                <Select.Content align="end">
                  <Select.Option className="text-base" value="bottom-right">
                    Bottom right
                  </Select.Option>
                  <Select.Option className="text-base" value="bottom-left">
                    Bottom left
                  </Select.Option>
                </Select.Content>
              </Select>
            </WidgetOptionRow>
          ) : null}
          {mode === "custom" ? (
            <WidgetOptionRow
              description="The loader opens when someone clicks an element matching this selector."
              label="Trigger selector"
            >
              <Input
                className="h-10 w-full sm:w-72"
                id="feedback-widget-trigger"
                onChange={(event) => {
                  setTrigger(event.target.value);
                }}
                value={trigger}
              />
            </WidgetOptionRow>
          ) : null}
        </Box>
      </Box>

      <Box className="border-border bg-surface mt-6 overflow-hidden rounded-2xl border">
        <SectionHeader
          description="Add this snippet to your product to install the feedback widget."
          title="Installation"
        />
        <Box className="p-6">
          <WidgetCodeBlock code={snippet} language="HTML" />
        </Box>
      </Box>
    </>
  );
};
