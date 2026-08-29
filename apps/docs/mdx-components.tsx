import defaultMdxComponents from "fumadocs-ui/mdx";
import { ImageZoom } from "fumadocs-ui/components/image-zoom";
import type { MDXComponents } from "mdx/types";

// use this function to get MDX components, you will need it for rendering MDX
export function getMDXComponents(components?: MDXComponents): MDXComponents {
  return {
    ...defaultMdxComponents,
    img: (props) => (
      <span className="relative block">
        <ImageZoom {...(props as any)} />
        <span className="pointer-events-none absolute inset-0 z-[3] block rounded-md bg-[url('/noise.png')] bg-repeat opacity-40" />
      </span>
    ),
    ...components,
  };
}
