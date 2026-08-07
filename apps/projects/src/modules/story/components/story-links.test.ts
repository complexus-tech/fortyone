/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("Story links", () => {
  it("keeps the add-link action available before a story has links", () => {
    const optionsSource = readSource(
      "src/modules/story/components/options.tsx",
    );
    const addLinksSource = readSource(
      "src/modules/story/components/add-links.tsx",
    );

    expect(optionsSource).toContain("<AddLinks storyId={storyId} />");
    expect(addLinksSource).toContain("Add link");
    expect(addLinksSource).toContain("<AddLinkDialog");
  });

  it("uses one flat divider treatment across story detail sections", () => {
    const linksSource = readSource("src/modules/story/components/links.tsx");
    const associationsSource = readSource(
      "src/modules/story/components/associations.tsx",
    );
    const subStoriesSource = readSource(
      "src/modules/story/components/sub-stories.tsx",
    );
    const relatedDocumentsSource = readSource(
      "src/modules/documents/related-documents.tsx",
    );

    expect(linksSource).not.toContain("border-t-[0.5px]");
    expect(associationsSource).not.toContain("border-t-[0.5px]");
    expect(subStoriesSource).not.toContain("border-t-[0.5px]");
    expect(relatedDocumentsSource).not.toContain("border-t");
  });

  it("keeps consistent spacing between story detail sections", () => {
    const linksSource = readSource("src/modules/story/components/links.tsx");
    const associationsSource = readSource(
      "src/modules/story/components/associations.tsx",
    );
    const subStoriesSource = readSource(
      "src/modules/story/components/sub-stories.tsx",
    );
    const relatedDocumentsSource = readSource(
      "src/modules/documents/related-documents.tsx",
    );
    const mainDetailsSource = readSource(
      "src/modules/story/components/main-details.tsx",
    );

    expect(linksSource).toContain('<Box className="mt-4">');
    expect(associationsSource).toContain('<Box className="mt-4">');
    expect(subStoriesSource).toContain('<Box className="mt-5">');
    expect(subStoriesSource).toContain('className="h-auto pt-1 pb-0"');
    expect(relatedDocumentsSource).toContain('<Box className="mt-4">');
    expect(mainDetailsSource).toContain(
      '<Attachments className="mt-4" storyId={storyId} />',
    );
    expect(mainDetailsSource).toContain('className={cn("max-w-5xl');
    expect(mainDetailsSource).toContain("-left-px mb-5 text-3xl");
  });

  it("renders related documents as link-style rows with document actions", () => {
    const source = readSource("src/modules/documents/related-documents.tsx");

    expect(source).toContain("<RowWrapper");
    expect(source).toContain("Open document");
    expect(source).toContain("Remove association");
    expect(source).toContain("<TimeAgo timestamp={document.updatedAt} />");
    expect(source).not.toContain("document.contentText");
  });
});
