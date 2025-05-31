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
        <span className="pointer-events-none block absolute inset-0 z-[3] bg-[url('/noise.png')] bg-repeat opacity-50 rounded-xl" />
      </span>
    ),
    ...components,
  };
}
