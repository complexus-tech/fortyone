import Link from "next/link";
import type { ArticleHeading } from "./article-outline";
import styles from "./article.module.css";

export function ArticleMap({ headings }: { headings: ArticleHeading[] }) {
  if (headings.length < 2) return null;
  return (
    <details className={styles.contents}>
      <summary>In this article</summary>
      <nav aria-label="Article sections">
        <ol>
          {headings.map((heading) => (
            <li key={heading.id}>
              <Link href={`#${heading.id}`}>{heading.title}</Link>
            </li>
          ))}
        </ol>
      </nav>
    </details>
  );
}
