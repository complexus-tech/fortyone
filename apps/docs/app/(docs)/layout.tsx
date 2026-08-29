import { baseOptions } from "@/app/layout.config";
import { FortyOneDocsLayout } from "@/components/docs-layout";
import { source } from "@/lib/source";
import type { ReactNode } from "react";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <FortyOneDocsLayout baseOptions={baseOptions} tree={source.getPageTree()}>
      {children}
    </FortyOneDocsLayout>
  );
}
