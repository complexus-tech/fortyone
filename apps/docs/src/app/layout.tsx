import { Layout, Navbar } from "nextra-theme-docs";
import { Head } from "nextra/components";
import { getPageMap } from "nextra/page-map";
import "nextra-theme-docs/style.css";
import type { Metadata } from "next";
import { Instrument_Sans as InstrumentSans } from "next/font/google";

import "./globals.css";
import { MainLogo } from "@/components/logo";
import { InternetIcon, SupportIcon } from "icons";

const font = InstrumentSans({
  subsets: ["latin"],
  display: "swap",
  weight: ["500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Complexus Docs",
  description: "Learn how to use Complexus to manage your projects.",
};

const navbar = (
  <Navbar
    logo={
      <MainLogo className="relative -left-3.5 top-0.5 z-10 h-5 md:h-[1.6rem]" />
    }
    chatLink="https://complexus.app/contact"
    chatIcon={<SupportIcon className="opacity-80 h-[1.1rem]" />}
    projectLink="https://complexus.app"
    projectIcon={<InternetIcon className="opacity-80 h-[1.1rem]" />}
  />
);

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      dir="ltr"
      suppressHydrationWarning
      className={font.className}
    >
      <Head
        backgroundColor={{
          dark: "#0d0e10",
          light: "#ffffff",
        }}
        color={{
          lightness: 65,
          saturation: 77,
          hue: 0,
        }}

        // ... Your additional head options
      >
        {/* Your additional tags should be passed as `children` of `<Head>` element */}
      </Head>
      <body className="antialiased">
        <Layout
          navbar={navbar}
          pageMap={await getPageMap()}
          editLink={null}
          sidebar={{
            defaultMenuCollapseLevel: 1,
          }}
        >
          {children}
        </Layout>
      </body>
    </html>
  );
}
