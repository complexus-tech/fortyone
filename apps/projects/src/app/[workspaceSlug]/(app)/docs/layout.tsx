import type { ReactNode } from "react";
import { Suspense } from "react";
import { DocumentsShell } from "@/modules/documents/documents-shell";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <Suspense fallback={<div className="h-dvh" />}>
      <DocumentsShell>{children}</DocumentsShell>
    </Suspense>
  );
}
