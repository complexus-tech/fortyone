type HtmlNode = {
  type: string;
  tagName?: string;
  value?: string;
  properties?: Record<string, unknown>;
  children?: HtmlNode[];
};

export type ArticleHeading = { id: string; title: string };

function nodeText(node: HtmlNode): string {
  return node.value ?? node.children?.map(nodeText).join("") ?? "";
}

/** Read IDs from the rendered tree, after rehype-slug has assigned them. */
export function createArticleOutline() {
  const headings: ArticleHeading[] = [];
  function visit(node: HtmlNode) {
    if (
      node.type === "element" &&
      node.tagName === "h2" &&
      typeof node.properties?.id === "string"
    ) {
      headings.push({ id: node.properties.id, title: nodeText(node) });
    }
    node.children?.forEach(visit);
  }
  return { headings, plugin: () => visit };
}
