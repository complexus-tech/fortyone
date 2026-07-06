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
  compiler: {
    removeConsole: process.env.NODE_ENV === "production",
  },
  skipTrailingSlashRedirect: true,
};

module.exports = nextConfig;
