import { icons } from "lucide-react";
import { loader } from "fumadocs-core/source";
import { defineDocs } from "fumadocs-mdx/macro";
import { createElement } from "react";

import { GoIcon, TypeScriptIcon } from "@/components/sdk-icons";
import { openapi } from "@/lib/openapi";

const docs = defineDocs({
  dir: "content/docs",
});

const customIcons = {
  Go: GoIcon,
  TypeScript: TypeScriptIcon,
} as const;

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
      if (icon && icon in customIcons) {
        return createElement(customIcons[icon as keyof typeof customIcons]);
      }
      if (icon && icon in icons) {
        return createElement(icons[icon as keyof typeof icons]);
      }
    },
  },
);
