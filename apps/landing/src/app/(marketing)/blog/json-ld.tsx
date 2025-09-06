import type { WithContext, Blog, BreadcrumbList } from "schema-dts";

const blogSchema: WithContext<Blog> = {
  "@context": "https://schema.org",
  "@type": "Blog",
  name: "Forty One Project Management Blog",
  description:
    "Expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  publisher: {
    "@type": "Organization",
    name: "Forty One",
    logo: "https://fortyone.app/images/logo.png",
  },
};

const breadcrumbSchema: WithContext<BreadcrumbList> = {
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  itemListElement: [
    {
      "@type": "ListItem",
      position: 1,
      item: {
        "@id": "https://fortyone.app",
        name: "Home",
      },
    },
    {
      "@type": "ListItem",
      position: 2,
      item: {
        "@id": "https://fortyone.app/blog",
        name: "Blog",
      },
    },
  ],
};

export const BlogJsonLd = () => {
  return (
    <>
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(blogSchema) }}
        type="application/ld+json"
      />
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbSchema) }}
        type="application/ld+json"
      />
    </>
  );
};
