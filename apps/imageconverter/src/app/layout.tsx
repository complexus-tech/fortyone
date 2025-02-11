import type { Metadata } from "next";
import type { ReactNode } from "react";
import "../styles/global.css";
import dynamic from "next/dynamic";
import { cn } from "lib";
import type { WebApplication, WithContext } from "schema-dts";
import { ThemeProvider } from "next-themes";
import { inter, satoshi } from "@/styles/fonts";
import { CursorProvider } from "@/context";
import { Footer, JsonLd, Navigation } from "@/components/shared";
import { PostHogProvider } from "./posthog";

const PostHogPageView = dynamic(() => import("./posthog-page-view"), {
  ssr: false,
});

export const metadata: Metadata = {
  metadataBase: new URL("https://imageconveta.com"),
  title: "Image Converter- Convert Your Images In Seconds",
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
    "jpg to webp",
    "png to jpg",
    "webp to png",
    "png to webp",
    "avif to png",
    "png to avif",
    "ai to png",
    "png to ai",
    "psd to png",
    "converter jpg to webp",
    "converter png to jpg",
    "converter png to webp",
    "converter webp to png",
    "converter avif to png",
    "converter ai to png",
    "converter psd to png",
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
      className={cn(satoshi.variable, inter.variable)}
      lang="en"
      suppressHydrationWarning
    >
      <body className="relative">
        <JsonLd>{structuredData}</JsonLd>
        <ThemeProvider attribute="class">
          <PostHogProvider>
            <CursorProvider>
              <Navigation />
              {children}
              <Footer />
            </CursorProvider>
            <PostHogPageView />
          </PostHogProvider>
        </ThemeProvider>
        <div className="pointer-events-none fixed left-0 top-0 h-full w-full bg-[url('/noise.png')] bg-repeat opacity-50" />
      </body>
    </html>
  );
}
