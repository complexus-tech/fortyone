"use client";

import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { Box, Input, Text } from "ui";
import { cn } from "lib";
import { SectionHeader } from "@/modules/settings/components";
import {
  buildFeedbackWidgetIdentityServerExample,
  buildFeedbackWidgetSnippet,
  buildInlineFeedbackWidgetMarkup,
} from "@/modules/feedback-widget/install";
import type {
  FeedbackWidgetMode,
  FeedbackWidgetTab,
  FeedbackWidgetTheme,
} from "@/modules/feedback-widget/protocol";
import { useFeedbackWidgetSettings } from "./hooks";
import { WidgetSecuritySettings } from "./widget-security-settings";

const selectClassName =
  "border-border bg-surface-elevated ring-ring h-10 w-full rounded-xl border px-3 text-sm outline-none focus-visible:ring-2";

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
  attribute: "text-[#d2a8ff]",
  comment: "text-[#8b949e]",
  keyword: "text-[#ff7b72]",
  number: "text-[#79c0ff]",
  plain: "text-[#e6edf3]",
  punctuation: "text-[#8b949e]",
  string: "text-[#a5d6ff]",
  tag: "text-[#7ee787]",
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

const copyText = async (value: string) => {
  await navigator.clipboard.writeText(value);
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
      await copyText(code);
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
    <figure className="overflow-hidden rounded-lg bg-[#1b1b19] text-[#e6edf3] shadow-sm ring-1 ring-black/10 dark:ring-white/10">
      <figcaption className="flex min-h-9 items-center justify-between bg-[#222220] px-3 font-mono text-[10px] tracking-[0.08em] text-[#8b949e] uppercase">
        <span>{language}</span>
        <button
          aria-live="polite"
          className="focus-visible:ring-ring rounded-sm py-2 text-inherit transition-colors hover:text-white focus-visible:ring-1 focus-visible:outline-none"
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
          "overflow-auto px-5 py-[18px] font-mono text-[12px] leading-[1.65] whitespace-pre",
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

export const WidgetInstallSettings = ({
  portalId,
  portalSlug,
  scriptOrigin,
}: {
  portalId: string;
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
  const widgetSettings = useFeedbackWidgetSettings(portalId);
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
    widgetKeyId: widgetSettings.data?.widgetKeyId,
  };
  const snippet =
    mode === "inline"
      ? buildInlineFeedbackWidgetMarkup(options)
      : buildFeedbackWidgetSnippet({ ...options, mode });
  const widgetConfiguration = widgetSettings.data;
  const identityServerExample =
    widgetConfiguration?.hasSigningSecret &&
    widgetConfiguration.allowedOrigins[0]
      ? buildFeedbackWidgetIdentityServerExample({
          origin: widgetConfiguration.allowedOrigins[0],
          signingSecretVersion: widgetConfiguration.signingSecretVersion,
          widgetKeyId: widgetConfiguration.widgetKeyId,
        })
      : "";

  return (
    <Box className="border-border bg-surface mb-6 overflow-hidden rounded-2xl border">
      <SectionHeader
        description="Add Feedback, Roadmap, and published Updates to your product without sending customers to a separate page."
        title="Feedback widget"
      />
      <Box className="space-y-6 p-6">
        <WidgetSecuritySettings portalId={portalId} />

        <Box className="border-border/70 border-t pt-6">
          <Text fontWeight="medium">Appearance and installation</Text>
          <Text className="mt-1 text-sm" color="muted">
            Configure how feedback opens in your product, then add the generated
            script to your site.
          </Text>
        </Box>
        <Box className="border-border/70 bg-background/60 grid gap-5 rounded-xl border p-5 sm:grid-cols-2 lg:grid-cols-4">
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
              <option value="updates">Updates</option>
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

        <Box>
          <Text fontWeight="medium">Add the widget to your product</Text>
          <Text className="mt-1 mb-4 max-w-2xl text-sm" color="muted">
            Paste this snippet before the closing body tag. The widget loads
            asynchronously and stays out of the way until someone opens it.
          </Text>
          <WidgetCodeBlock code={snippet} language="HTML" />
        </Box>
        {widgetSettings.data?.hasSigningSecret ? (
          <Box className="border-border/70 bg-background/60 rounded-xl border p-5">
            <Text fontWeight="medium">Identify signed-in customers</Text>
            <Text className="mt-1 max-w-2xl text-sm leading-5" color="muted">
              Generate a short-lived assertion on your backend with the widget
              key ID and signing secret, return only that assertion to your
              page, then pass it to the loader. Assertions stay in memory and
              are never added to the iframe URL or browser storage.
            </Text>
            <Box className="mt-4">
              <WidgetCodeBlock
                code={identityServerExample}
                language="JavaScript"
              />
            </Box>
            <Text className="mt-3 text-sm" color="muted">
              Then pass the assertion returned by your authenticated backend to
              the loader:
            </Text>
            <Box className="mt-3">
              <WidgetCodeBlock
                code={`// signedAssertion comes from your authenticated backend
window.FortyOneFeedback.identify(signedAssertion);`}
                language="JavaScript"
                maxHeight="max-h-40"
              />
            </Box>
          </Box>
        ) : null}
      </Box>
    </Box>
  );
};
