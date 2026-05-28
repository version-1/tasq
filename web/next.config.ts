import type { NextConfig } from "next";

const issueTrackerURL =
  process.env.ISSUE_TRACKER_INTERNAL_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${issueTrackerURL}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
