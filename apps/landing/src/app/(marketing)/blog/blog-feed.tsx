import type { ReactNode } from "react";
import { Container } from "@/components/ui";
import type { PostSummary } from "./post-card";
import { PostCard } from "./post-card";

export function BlogFeed({
  posts,
  children,
}: {
  posts: PostSummary[];
  children: ReactNode;
}) {
  const [featured, ...remaining] = posts;
  return (
    <Container>
      {children}
      {featured ? <PostCard featured post={featured} /> : null}
      {remaining.length > 0 ? (
        <section
          aria-label="More from the blog"
          className="border-border mt-12 border-t pt-12 md:mt-14 md:pt-14"
        >
          <div className="grid gap-x-8 gap-y-12 sm:grid-cols-2 lg:grid-cols-3">
            {remaining.map((post) => (
              <PostCard key={post.slug} post={post} />
            ))}
          </div>
        </section>
      ) : null}
    </Container>
  );
}
