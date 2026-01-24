import createMDX from "@next/mdx";

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ["ui", "icons", "next-mdx-remote"],
  pageExtensions: ["tsx", "ts", "mdx"],
  reactCompiler: true,
  experimental: {
    turbopackFileSystemCacheForDev: true,
    mdxRs: true,
    staleTimes: {
      dynamic: 30,
      static: 180,
    },
  },
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "lh3.googleusercontent.com",
      },
      {
        protocol: "https",
        hostname: "images.unsplash.com",
      },
    ],
  },
  async redirects() {
    return [
      {
        source: "/features",
        destination: "/features/tasks",
        permanent: true,
      },
      {
        source: "/features/stories",
        destination: "/features/tasks",
        permanent: true,
      },
      {
        source: "/product/stories",
        destination: "/features/tasks",
        permanent: true,
      },
      {
        source: "/product/:path*",
        destination: "/features/:path*",
        permanent: true,
      },
      {
        source: "/login",
        destination: "https://cloud.fortyone.app/",
        permanent: true,
      },
      {
        source: "/signup",
        destination: "https://cloud.fortyone.app/signup",
        permanent: true,
      },
      {
        source: "/auth-callback",
        destination: "https://cloud.fortyone.app/auth-callback",
        permanent: true,
      },
      {
        source: "/verify/:email/:token",
        destination: "https://cloud.fortyone.app/verify/:email/:token",
        permanent: true,
      },
      {
        source: "/onboarding/:path*",
        destination: "https://cloud.fortyone.app/onboarding/:path*",
        permanent: true,
      },
    ];
  },
  async rewrites() {
    return [
      {
        source: "/ingest/static/:path*",
        destination: "https://us-assets.i.posthog.com/static/:path*",
      },
      {
        source: "/ingest/:path*",
        destination: "https://us.i.posthog.com/:path*",
      },
      {
        source: "/ingest/decide",
        destination: "https://us.i.posthog.com/decide",
      },
    ];
  },
};

const withMDX = createMDX({});

export default withMDX(nextConfig);
