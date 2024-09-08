import type { Metadata } from "next";
import type { ReactNode } from "react";
import { Footer, JsonLd, Navigation } from "@/components/shared";
import "../styles/global.css";
import { CursorProvider } from "@/context";
import { PostHogProvider } from "./posthog";
import dynamic from "next/dynamic";
import { inter, satoshi } from "@/styles/fonts";
import { cn } from "lib";
import { WebApplication, WithContext } from "schema-dts";

const PostHogPageView = dynamic(() => import("./posthog-page-view"), {
  ssr: false,
});

export const metadata: Metadata = {
  title: "Image Converter - Convert Your Images In Seconds",
  keywords: [
    "image converter",
    "image converter online",
    "image converter free",
    "image converter tool",
    "image converter website",
    "image converter app",
    "image converter software",
    "image converter png",
    "image converter jpg",
    "image converter webp",
    "image converter avif",
    "image converter ai",
    "image converter psd",
    "jpg to png",
    "png to jpg",
    "webp to png",
    "png to webp",
    "avif to png",
    "png to avif",
    "ai to png",
    "png to ai",
    "psd to png",
  ],
  description:
    "Convert, resize, and optimize images effortlessly with ImageConveta. Perfect for designers, marketers, and developers.",
  openGraph: {
    type: "website",
    locale: "en_US",
    url: "https://imageconveta.com",
    siteName: "ImageConveta",
    title: "Image Converter - Convert Your Images In Seconds",
    description:
      "Convert, resize, and optimize images effortlessly with ImageConveta. Perfect for designers, marketers, and developers.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Image Converter - Convert Your Images In Seconds",
    description:
      "Convert, resize, and optimize images effortlessly with ImageConveta. Perfect for designers, marketers, and developers.",
  },
};

const structuredData: WithContext<WebApplication> = {
  "@context": "https://schema.org",
  "@type": "WebApplication",
  name: "ImageConveta",
  image: "https://imageconveta.com/converter.png",
  description:
    "Convert, resize, and optimize images effortlessly with ImageConveta. Perfect for designers, marketers, and developers.",
  applicationCategory: "ImageConversion",
  applicationSubCategory: "ImageConversion",
  inLanguage: "en",
  url: "https://imageconveta.com",
  featureList: [
    "Convert images to PNG, JPG, WEBP, AVIF, PDF, SVG, AI, PSD, and more.",
    "Resize images to any size.",
    "Optimize images for web.",
    "Batch process images.",
    "Free and easy to use.",
    "API available for developers.",
  ],
  creator: {
    "@type": "Organization",
    name: "ImageConveta",
    url: "https://imageconveta.com",
    logo: "https://imageconveta.com/logo.png",
  },
};

export default function RootLayout({
  children,
}: {
  children: ReactNode;
}): JSX.Element {
  return (
    <html
      lang="en"
      className={cn(satoshi.variable, inter.variable)}
      suppressHydrationWarning
    >
      <body className="relative">
        <JsonLd>{structuredData}</JsonLd>
        <PostHogProvider>
          <CursorProvider>
            <Navigation />
            {children}
            <Footer />
          </CursorProvider>
          <PostHogPageView />
        </PostHogProvider>
        <div className="pointer-events-none fixed left-0 top-0 h-full w-full bg-[url('/noise.png')] bg-repeat opacity-50" />
      </body>
    </html>
  );
}
