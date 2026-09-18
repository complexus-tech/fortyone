"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { cn } from "lib";
import { PublicBranding } from "@/components/ui/public-branding";
import { DocumentIndex } from "./document-index";
import styles from "./document-index.module.css";

export function PublicDocumentLayout({ children }: { children: ReactNode }) {
  const [scrollContainer, setScrollContainer] = useState<HTMLDivElement | null>(
    null,
  );
  return (
    <main
      className={cn(
        "bg-background text-foreground relative h-dvh",
        styles.host,
      )}
    >
      <div className="h-full overflow-y-auto" ref={setScrollContainer}>
        <div className={styles.contentPadding}>{children}</div>
      </div>
      <DocumentIndex
        contentSelector=".rich-document-editor"
        scrollContainer={scrollContainer}
      />
      <PublicBranding />
    </main>
  );
}
