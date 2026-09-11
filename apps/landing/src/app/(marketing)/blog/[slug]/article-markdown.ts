import type { MDXRemoteProps } from "next-mdx-remote/rsc";
import rehypeAutolinkHeadings from "rehype-autolink-headings";
import rehypePrettyCode from "rehype-pretty-code";
import rehypeSlug from "rehype-slug";
import remarkGfm from "remark-gfm";
import { createArticleOutline } from "./article-outline.ts";

export function createArticleMarkdown() {
  const outline = createArticleOutline();
  const options: MDXRemoteProps["options"] = {
    mdxOptions: {
      remarkPlugins: [remarkGfm],
      rehypePlugins: [
        rehypeSlug,
        outline.plugin,
        [
          rehypeAutolinkHeadings,
          {
            behavior: "wrap",
            properties: { className: ["heading-anchor"] },
          },
        ],
        [
          rehypePrettyCode,
          {
            theme: {
              light: "github-light-default",
              dark: "github-dark-default",
            },
            keepBackground: false,
            bypassInlineCode: true,
            defaultLang: "text",
          },
        ],
      ],
    },
  };
  return { headings: outline.headings, options };
}
