import { openapi } from "@/lib/openapi";
import { icons } from "lucide-react";
import { loader } from "fumadocs-core/source";
import { defineDocs } from "fumadocs-mdx/macro";
import { createElement } from "react";

const docs = defineDocs({
  dir: "content/docs",
});

const apiReference = await openapi.staticSource({
  baseDir: "api-reference/reference",
  groupBy: "tag",
  meta: {
    folderStyle: "folder",
  },
});

// See https://fumadocs.vercel.app/docs/headless/source-api for more info
export const source = loader(
  {
    apiReference,
    docs: docs.toFumadocsSource(),
  },
  {
    baseUrl: "/",
    plugins: [openapi.loaderPlugin()],
    icon(icon) {
      if (icon && icon in icons) {
        return createElement(icons[icon as keyof typeof icons]);
      }
    },
  },
);
