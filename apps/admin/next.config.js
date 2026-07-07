const path = require("node:path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: "standalone",
  transpilePackages: ["ui", "icons"],
  devIndicators: false,
  reactCompiler: true,
  experimental: {
    turbopackFileSystemCacheForDev: false,
    staleTimes: {
      dynamic: 30,
      static: 180,
    },
    serverActions: {
      bodySizeLimit: "5mb",
    },
  },
  turbopack: {
    root: path.join(__dirname, "../.."),
  },
  compiler: {
    removeConsole: process.env.NODE_ENV === "production",
  },
  images: {
    dangerouslyAllowLocalIP: true,
    remotePatterns: [
      {
        protocol: "http",
        hostname: "minio",
        port: "9000",
      },
      {
        protocol: "http",
        hostname: "localhost",
        port: "9000",
      },
      {
        protocol: "https",
        hostname: "lh3.googleusercontent.com",
      },
      {
        protocol: "https",
        hostname: "avatars.githubusercontent.com",
      },
      {
        protocol: "https",
        hostname: "user-images.githubusercontent.com",
      },
      {
        protocol: "https",
        hostname: "fortyone.s3.us-east-1.amazonaws.com",
      },
      {
        protocol: "https",
        hostname: "images.unsplash.com",
      },
      {
        protocol: "https",
        hostname: "complexus.blob.core.windows.net",
      },
      {
        protocol: "https",
        hostname: "www.fortyone.app",
      },
      {
        protocol: "https",
        hostname: "www.fortyone.lc",
      },
    ],
  },
  skipTrailingSlashRedirect: true,
};

module.exports = nextConfig;
