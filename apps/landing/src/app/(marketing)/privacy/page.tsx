import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { Box, Container } from "ui";
import { notFound } from "next/navigation";
import { Markdown } from "@/components/markdown";

export const metadata = {
  title: "Privacy Policy | FortyOne",
  description: "Learn how FortyOne protects your data and privacy.",
};

async function getPrivacyPolicy() {
  try {
    const filePath = join(process.cwd(), "src/content/privacy-policy.md");
    const content = await readFile(filePath, "utf8");
    return content.replace("{{date}}", new Date().toLocaleDateString());
  } catch (error) {
    notFound();
  }
}

export default async function PrivacyPage() {
  const content = await getPrivacyPolicy();

  return (
    <Container className="py-24 md:pt-36">
      <Box className="prose prose-lg prose-stone mx-auto max-w-3xl px-1 leading-7 dark:prose-invert prose-headings:font-medium prose-a:text-primary prose-pre:text-[1.1rem] prose-pre:bg-surface-muted prose-pre:text-foreground">
        <Markdown content={content} />
      </Box>
    </Container>
  );
}
