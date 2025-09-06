import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { Box, Container } from "ui";
import { notFound } from "next/navigation";
import { Markdown } from "@/components/markdown";

export const metadata = {
  title: "Terms of Service | FortyOne",
  description:
    "Read our Terms of Service to understand the rules and guidelines for using FortyOne.",
};

async function getTerms() {
  try {
    const filePath = join(process.cwd(), "src/content/terms.md");
    const content = await readFile(filePath, "utf8");
    return content.replace("{{date}}", new Date().toLocaleDateString());
  } catch (error) {
    notFound();
  }
}

export default async function TermsPage() {
  const content = await getTerms();

  return (
    <Container className="py-24 md:pt-36">
      <Box className="prose prose-lg prose-stone mx-auto max-w-3xl px-1 leading-7 dark:prose-invert prose-headings:font-medium prose-a:text-primary prose-pre:bg-gray-50 prose-pre:text-[1.1rem] prose-pre:text-dark-200 dark:prose-pre:bg-dark-200/80 dark:prose-pre:text-gray-200">
        <Markdown content={content} />
      </Box>
    </Container>
  );
}
