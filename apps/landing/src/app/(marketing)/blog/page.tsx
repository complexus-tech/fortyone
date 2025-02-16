import type { Metadata } from "next";
import { Box } from "ui";
import { ComingSoon } from "@/components/ui";
import { BlogJsonLd } from "./json-ld";

export const metadata: Metadata = {
  title: "Project Management Resources & Guides | Complexus Blog",
  description:
    "Access expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  keywords: [
    "project management blog",
    "OKR guides",
    "team collaboration tips",
    "project management resources",
    "agile project management",
    "OKR implementation",
    "team productivity tips",
    "project management best practices",
  ],
  openGraph: {
    title: "Project Management Resources & Guides | Complexus Blog",
    description:
      "Access expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  },
  twitter: {
    title: "Project Management Resources & Guides | Complexus Blog",
    description:
      "Access expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  },
};

export default function Page() {
  return (
    <>
      <BlogJsonLd />
      <Box className="pt-16 md:pt-0">
        <ComingSoon />
      </Box>
    </>
  );
}
