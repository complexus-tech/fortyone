import type { ComponentPropsWithoutRef, ReactNode } from "react";
import type { MDXComponents } from "mdx/types";
import Image from "next/image";
import { mdxComponents } from "@/mdx-components";
import { ArticleCode } from "./article-code";
import styles from "./article-components.module.css";

function Takeaway({
  children,
  title = "The takeaway",
}: {
  children: ReactNode;
  title?: string;
}) {
  return (
    <aside className={styles.takeaway}>
      <p className={styles.label}>{title}</p>
      <div>{children}</div>
    </aside>
  );
}

function ProductFigure({
  light,
  dark,
  caption,
  alt,
}: {
  light: string;
  dark: string;
  caption: string;
  alt: string;
}) {
  return (
    <figure className={styles.figure}>
      {[
        { theme: "light", src: light },
        { theme: "dark", src: dark },
      ].map(({ theme, src }) => (
        <a
          aria-label={`Enlarge image: ${alt}`}
          className={`${styles.productImage} rounded-2xl`}
          data-theme={theme}
          href={src}
          key={theme}
          rel="noopener noreferrer"
          target="_blank"
        >
          <Image
            alt={alt}
            height={1790}
            sizes="(max-width: 1023px) 100vw, 820px"
            src={src}
            width={3000}
          />
        </a>
      ))}
      <figcaption>{caption}</figcaption>
    </figure>
  );
}

function WorkflowSteps({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <figure className={styles.workflow}>
      <figcaption className={styles.label}>{title}</figcaption>
      <ol className={styles.steps}>{children}</ol>
    </figure>
  );
}

function WorkflowStep({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <li>
      <span aria-hidden="true" className={styles.stepNumber} />
      <strong>{title}</strong>
      <span>{children}</span>
    </li>
  );
}

function ArticleChecklist({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <aside className={styles.checklist}>
      <p className={styles.label}>{title}</p>
      <ul>{children}</ul>
    </aside>
  );
}

function ChecklistItem({ children }: { children: ReactNode }) {
  return <li>{children}</li>;
}

function ArticleDetails({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <details className={styles.details}>
      <summary>{title}</summary>
      <div>{children}</div>
    </details>
  );
}

/* eslint-disable jsx-a11y/no-noninteractive-tabindex -- Wide article tables need a keyboard-focusable scroll region. */
function ArticleTable({
  children,
  ...props
}: ComponentPropsWithoutRef<"table">) {
  return (
    <div
      aria-label="Article table"
      className={styles.tableScroll}
      role="region"
      tabIndex={0}
    >
      <table {...props}>{children}</table>
    </div>
  );
}

/* eslint-enable jsx-a11y/no-noninteractive-tabindex -- Restore the rule outside the scroll region. */

export const articleMdxComponents: MDXComponents = {
  ...mdxComponents,
  code: ({ children, ...props }: ComponentPropsWithoutRef<"code">) => (
    <code {...props}>{children}</code>
  ),
  pre: ArticleCode,
  table: ArticleTable,
  Takeaway,
  ProductFigure,
  WorkflowSteps,
  WorkflowStep,
  ArticleChecklist,
  ChecklistItem,
  ArticleDetails,
};
