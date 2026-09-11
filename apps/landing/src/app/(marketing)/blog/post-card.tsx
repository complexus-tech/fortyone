import Image from "next/image";
import Link from "next/link";
import { ArrowUpRightIcon } from "icons";
import { cn } from "lib";
import type { PostMetadata } from "@/lib/posts";
import styles from "./blog-list.module.css";

export type PostSummary = {
  slug: string;
  metadata: PostMetadata;
  readingTime: number;
};

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  day: "numeric",
  month: "short",
  year: "numeric",
  timeZone: "UTC",
});

export function PostMeta({
  post,
  appearance = "badge",
}: {
  post: PostSummary;
  appearance?: "badge" | "plain";
}) {
  return (
    <div className="text-text-muted flex flex-wrap items-center gap-x-3 gap-y-2 text-sm">
      <span
        className={
          appearance === "badge"
            ? cn(styles.category, "rounded-md")
            : "text-foreground"
        }
        data-category={post.metadata.category}
      >
        {post.metadata.category}
      </span>
      <time dateTime={post.metadata.date}>
        {dateFormatter.format(new Date(post.metadata.date))}
      </time>
    </div>
  );
}

export function PostCard({
  post,
  featured = false,
}: {
  post: PostSummary;
  featured?: boolean;
}) {
  return (
    <article className="min-w-0">
      <Link
        className={cn(
          styles.postLink,
          !featured && "flex h-full flex-col",
          "group focus-visible:outline-ring block rounded-2xl focus-visible:outline-2 focus-visible:outline-offset-8",
          featured && styles.featured,
        )}
        href={`/blog/${post.slug}`}
      >
        <div
          className={cn(
            "bg-surface-muted relative aspect-[1.85/1] overflow-hidden rounded-2xl sm:rounded-[2rem]",
          )}
        >
          <Image
            alt=""
            className="object-cover transition-transform duration-500 ease-out group-hover:scale-[1.025] motion-reduce:transition-none motion-reduce:group-hover:scale-100"
            fill
            priority={featured}
            sizes={
              featured
                ? "(max-width: 767px) 100vw, 50vw"
                : "(max-width: 639px) 100vw, (max-width: 1023px) 50vw, 33vw"
            }
            src={post.metadata.featuredImage}
          />
        </div>
        <div
          className={featured ? "py-2 md:py-6" : "flex flex-1 flex-col pt-6"}
        >
          <PostMeta post={post} />
          <h2
            className={cn(
              styles.title,
              "mt-4 text-pretty",
              featured
                ? "font-serif text-3xl leading-[1.12] font-normal md:text-4xl"
                : "text-xl font-semibold",
            )}
          >
            {post.metadata.title}
          </h2>
          <p
            className={cn(
              "text-text-description mt-3 leading-relaxed",
              featured ? "text-base" : "text-sm",
            )}
          >
            {post.metadata.description}
          </p>
          <div className="text-text-muted mt-auto flex items-center justify-between gap-4 pt-6 text-sm">
            <span>
              {post.metadata.author}
              <span aria-hidden="true"> · </span>
              {post.readingTime} min read
            </span>
            <ArrowUpRightIcon className="text-foreground size-5 shrink-0" />
          </div>
        </div>
      </Link>
    </article>
  );
}
