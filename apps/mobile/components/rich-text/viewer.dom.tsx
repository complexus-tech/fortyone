"use dom";

import { useEffect, useMemo, useRef } from "react";
import { colors, themeColors } from "../../constants/colors";
import type { DOMProps } from "expo/dom";
import { sanitizeRichText } from "./sanitize";

type Props = {
  html: string;
  dark: boolean;
  onOpenLink: (href: string) => Promise<void>;
  dom?: DOMProps;
};

export default function RichTextViewerDOM({ html, dark, onOpenLink }: Props) {
  const sanitized = useMemo(() => sanitizeRichText(html), [html]);
  const articleRef = useRef<HTMLElement>(null);
  const theme = themeColors[dark ? "dark" : "light"];
  useEffect(() => {
    // Restore only schema-owned media presentation after sanitizing all user styles.
    for (const media of articleRef.current?.querySelectorAll<HTMLElement>(
      "img,video",
    ) ?? []) {
      const width = media.dataset.width ?? "100%";
      const alignment = media.dataset.align ?? "center";
      media.style.display = "block";
      media.style.width = /^\d+(?:\.\d+)?(?:px|%)$/.test(width)
        ? width
        : "100%";
      media.style.marginLeft = alignment === "left" ? "0" : "auto";
      media.style.marginRight = alignment === "right" ? "0" : "auto";
    }
  }, [sanitized]);
  return (
    <>
      <style>{`
      *{box-sizing:border-box}html,body{margin:0;padding:0;background:transparent;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}
      article{--line:${theme.border};--surface:${theme.surfaceMuted};--accent:${dark ? colors.primary : `color-mix(in srgb,${colors.primary} 75%,${colors.black})`};font-size:16px;line-height:1.6;color:${theme.foreground};overflow-wrap:anywhere;padding:1px 0}p{margin:.6em 0}h1,h2,h3{line-height:1.25}a{color:var(--accent)}img,video{max-width:100%;height:auto;border-radius:12px}pre{white-space:pre-wrap;background:var(--surface);padding:12px;border-radius:10px}code{font-size:.85em}blockquote{margin:12px 0;border-left:3px solid var(--line);padding-left:14px}table{border-collapse:collapse;width:100%;table-layout:fixed}th,td{border:1px solid var(--line);padding:7px;vertical-align:top}th{background:var(--surface)}ul,ol{padding-left:24px}ul[data-type=taskList]{list-style:none;padding-left:0}li[data-type=taskItem]{display:flex;gap:8px}li[data-type=taskItem]>label{padding-top:8px}li[data-type=taskItem]>div{flex:1;min-width:0}li p{margin:3px 0}input[type=checkbox]{accent-color:var(--accent);pointer-events:none;width:17px;height:17px}hr{border:0;border-top:1px solid var(--line)}
    `}</style>
      <article
        ref={articleRef}
        onClick={(event) => {
          const anchor =
            event.target instanceof Element ? event.target.closest("a") : null;
          if (anchor) {
            event.preventDefault();
            void onOpenLink(anchor.getAttribute("href") ?? "");
          }
        }}
        dangerouslySetInnerHTML={{ __html: sanitized }}
      />
    </>
  );
}
