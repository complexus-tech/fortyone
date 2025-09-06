import type { ContactPage, WithContext } from "schema-dts";

const contactPage: WithContext<ContactPage> = {
  "@context": "https://schema.org",
  "@type": "ContactPage",
  name: "Contact Forty One",
  description:
    "Contact page for Forty One - Modern OKR & Project Management Platform",
  mainEntity: {
    "@type": "Organization",
    name: "Forty One",
    url: "https://fortyone.app",
    logo: "https://fortyone.app/images/logo.png",
    contactPoint: {
      "@type": "ContactPoint",
      contactType: "customer support",
      email: "support@complexus.app",
      availableLanguage: ["English"],
      contactOption: "TollFree",
      areaServed: "Worldwide",
    },
  },
};

export const ContactJsonLd = () => {
  return (
    <script
      dangerouslySetInnerHTML={{ __html: JSON.stringify(contactPage) }}
      type="application/ld+json"
    />
  );
};
