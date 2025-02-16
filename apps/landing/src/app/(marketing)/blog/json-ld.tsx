import type { WithContext, Blog, BreadcrumbList } from "schema-dts";

const blogSchema: WithContext<Blog> = {
  "@context": "https://schema.org",
  "@type": "Blog",
  name: "Complexus Project Management Blog",
  description:
    "Expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  publisher: {
    "@type": "Organization",
    name: "Complexus",
    logo: "https://complexus.app/images/logo.png",
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
        "@id": "https://complexus.app",
        name: "Home",
      },
    },
    {
      "@type": "ListItem",
      position: 2,
      item: {
        "@id": "https://complexus.app/blog",
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
