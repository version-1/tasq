import type { Project } from "@/lib/generated/issue-tracker";

export const projectFixtures: Project[] = [
  {
    id: 1,
    key: "tasq",
    name: "Tasq",
    description: "Local issue tracking workspace for standalone web development.",
    location: "/workspace/tasq",
    createdAt: "2026-06-01T00:00:00.000Z",
    updatedAt: "2026-06-01T00:00:00.000Z",
  },
  {
    id: 2,
    key: "agent-docs",
    name: "agent-docs",
    description: "Documentation for agent development and usage.",
    location: "/workspace/agent-docs",
    createdAt: "2026-06-01T00:00:00.000Z",
    updatedAt: "2026-06-01T00:00:00.000Z",
  },
  {
    id: 3,
    key: "some-project",
    name: "SomeProject",
    description: "Some project description.",
    location: "/workspace/some-project",
    createdAt: "2026-06-01T00:00:00.000Z",
    updatedAt: "2026-06-01T00:00:00.000Z",
  },
];
