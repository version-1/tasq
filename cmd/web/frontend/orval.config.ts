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
  orchestrator: {
    input: {
      target: "../../../docs/openapi/orchestrator.yml",
    },
    output: {
      baseUrl: "/orchestrator",
      clean: false,
      client: "fetch",
      target: "src/lib/generated/orchestrator.ts",
    },
  },
});
