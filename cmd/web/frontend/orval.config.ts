import { defineConfig } from "orval";

export default defineConfig({
  issueTracker: {
    input: {
      target: "../../../docs/openapi/issue-tracker.yml",
    },
    output: {
      baseUrl: "/tracker",
      clean: true,
      client: "fetch",
      target: "src/lib/generated/issue-tracker.ts",
    },
  },
});
