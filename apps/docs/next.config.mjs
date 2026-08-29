import { createMDX } from "fumadocs-mdx/next";

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  reactStrictMode: true,
  experimental: {
    useTypeScriptCli: true,
  },
  async redirects() {
    return [
      {
        source: "/product-guide/stories",
        destination: "/product-guide/tasks",
        permanent: true,
      },
      {
        source: "/product-guide/stories/:path*",
        destination: "/product-guide/tasks/:path*",
        permanent: true,
      },
      {
        source: "/stories/:path*",
        destination: "/tasks/:path*",
        permanent: true,
      },
      {
        source: "/api",
        destination: "/api-reference",
        permanent: true,
      },
      {
        source: "/api/:path*",
        destination: "/api-reference/:path*",
        permanent: true,
      },
    ];
  },
};

export default withMDX(config);
