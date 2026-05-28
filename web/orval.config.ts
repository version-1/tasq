import { defineConfig } from "orval";

export default defineConfig({
  issueTracker: {
    input: {
      target: "../docs/openapi/issue-tracker.yml",
    },
    output: {
      baseUrl: '${process.env.NEXT_PUBLIC_ISSUE_TRACKER_URL ?? ""}',
      clean: true,
      client: "fetch",
      target: "lib/generated/issue-tracker.ts",
    },
  },
});
