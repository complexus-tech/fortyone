"use client";

import { useMemo, useState } from "react";
import { Button, Flex, Menu, Text, Tooltip } from "ui";
import {
  CodeIcon,
  CopyIcon,
  DocsIcon,
  Download04Icon,
  ExternalLinkIcon,
  MoreHorizontalIcon,
  PrintIcon,
} from "icons";
import { cn } from "lib";
import { Logo } from "@/components/ui/logo";
import { DocumentIndex } from "./document-index";
import { createStandaloneDocumentHTML } from "./public-document-export";
import styles from "./document-index.module.css";

type PublicDocumentLayoutProps = {
  contentHtml: string;
  markdown: string;
  title: string;
  updatedAt: string;
};

const safeFilename = (title: string) =>
  title
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "") || "fortyone-document";

const downloadTextFile = (contents: string, filename: string, type: string) => {
  const url = URL.createObjectURL(new Blob([contents], { type }));
  const anchor = window.document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
};

export function PublicDocumentLayout({
  contentHtml,
  markdown,
  title,
  updatedAt,
}: PublicDocumentLayoutProps) {
  const [scrollContainer, setScrollContainer] = useState<HTMLDivElement | null>(
    null,
  );
  const [announcement, setAnnouncement] = useState("");
  const filename = useMemo(() => safeFilename(title), [title]);
  const standaloneHTML = useMemo(
    () => createStandaloneDocumentHTML({ contentHtml, title, updatedAt }),
    [contentHtml, title, updatedAt],
  );
  const markdownDocument = `# ${title}\n\n_Updated ${updatedAt}_\n\n${markdown}`;

  const copyLink = async () => {
    await navigator.clipboard.writeText(window.location.href);
    setAnnouncement("Link copied");
  };

  return (
    <main
      className={cn(
        "bg-surface-muted text-foreground h-dvh overflow-hidden p-2 sm:p-3 print:h-auto print:overflow-visible print:bg-white print:p-0 print:text-black",
        styles.host,
      )}
    >
      <section className="border-border/80 bg-background shadow-shadow mx-auto flex h-[calc(100dvh-1rem)] max-w-[96rem] flex-col overflow-hidden rounded-xl border-[0.5px] shadow-sm sm:h-[calc(100dvh-1.5rem)] print:h-auto print:max-w-none print:overflow-visible print:rounded-none print:border-0 print:bg-white print:shadow-none">
        <header className="border-border/70 bg-background/95 sticky top-0 z-40 flex h-14 shrink-0 items-center justify-between border-b px-3 backdrop-blur-xl sm:px-4 print:hidden">
          <Flex align="center" className="min-w-0" gap={2}>
            <DocsIcon className="text-info size-4.5 shrink-0" />
            <Text className="truncate" fontWeight="medium">
              {title}
            </Text>
          </Flex>
          <Flex align="center" className="shrink-0" gap={1}>
            <Menu>
              <Tooltip title="Download">
                <Menu.Button>
                  <Button
                    aria-label="Download document"
                    asIcon
                    color="tertiary"
                    size="sm"
                    variant="naked"
                  >
                    <Download04Icon className="size-4.5" />
                  </Button>
                </Menu.Button>
              </Tooltip>
              <Menu.Items align="end" className="w-52 p-1.5">
                <Menu.Group className="px-0">
                  <Menu.Item
                    onSelect={() => {
                      window.print();
                    }}
                  >
                    <PrintIcon className="size-4" />
                    PDF
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      downloadTextFile(
                        standaloneHTML,
                        `${filename}.html`,
                        "text/html;charset=utf-8",
                      );
                    }}
                  >
                    <CodeIcon className="size-4" />
                    HTML
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      downloadTextFile(
                        markdownDocument,
                        `${filename}.md`,
                        "text/markdown;charset=utf-8",
                      );
                    }}
                  >
                    <DocsIcon className="size-4" />
                    Markdown
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
            <Menu>
              <Tooltip title="More options">
                <Menu.Button>
                  <Button
                    aria-label="More document options"
                    asIcon
                    color="tertiary"
                    size="sm"
                    variant="naked"
                  >
                    <MoreHorizontalIcon className="size-4.5" />
                  </Button>
                </Menu.Button>
              </Tooltip>
              <Menu.Items align="end" className="w-52 p-1.5">
                <Menu.Group className="px-0">
                  <Menu.Item onSelect={() => void copyLink()}>
                    <CopyIcon className="size-4" />
                    Copy link
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      window.print();
                    }}
                  >
                    <PrintIcon className="size-4" />
                    Print document
                  </Menu.Item>
                  <Menu.Separator />
                  <Menu.Item asChild>
                    <a href="https://fortyone.app">
                      <ExternalLinkIcon className="size-4" />
                      What is FortyOne?
                    </a>
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
            <span className="bg-border mx-1 hidden h-6 w-px sm:block" />
            <Button
              className="hidden sm:flex"
              color="tertiary"
              href="https://fortyone.app"
              size="sm"
              variant="outline"
            >
              Get FortyOne Free
            </Button>
          </Flex>
        </header>

        <div
          className="relative min-h-0 flex-1 overflow-y-auto overscroll-contain print:overflow-visible"
          ref={setScrollContainer}
        >
          <DocumentIndex
            contentSelector=".rich-document-editor"
            readingOffset={88}
            scrollContainer={scrollContainer}
          />
          <article className="mx-auto max-w-3xl px-6 pt-12 pb-24 sm:px-10 md:pt-16 lg:pt-20 print:max-w-none print:px-0 print:pt-6 print:pb-10">
            <h1 className="mb-3 text-4xl leading-[1.08] font-semibold tracking-[-0.035em] md:text-5xl">
              {title}
            </h1>
            <p className="text-text-muted mb-12 text-[0.95rem]">
              Last updated {updatedAt}
            </p>
            <div
              className="rich-document-editor rich-text-editor prose prose-lg prose-headings:tracking-[-0.02em] prose-img:max-w-full max-w-none print:text-black"
              dangerouslySetInnerHTML={{ __html: contentHtml }}
            />
          </article>
          <footer className="border-border/70 mx-auto flex max-w-3xl items-center justify-between border-t px-6 py-6 sm:px-10 print:hidden">
            <Flex align="center" className="text-text-muted" gap={2}>
              <Logo asIcon className="text-foreground h-3.5! w-auto!" />
              <Text color="muted" fontSize="sm">
                Published with FortyOne
              </Text>
            </Flex>
            <Button
              className="sm:hidden"
              color="tertiary"
              href="https://fortyone.app"
              size="sm"
              variant="outline"
            >
              Get FortyOne Free
            </Button>
          </footer>
        </div>
      </section>
      <span aria-live="polite" className="sr-only">
        {announcement}
      </span>
    </main>
  );
}
