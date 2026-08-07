import type { Metadata } from "next";
import { DocumentsHome } from "@/modules/documents/documents-home";

export const metadata: Metadata = { title: "Documents" };

export default function Page() {
  return <DocumentsHome />;
}
