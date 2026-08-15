import type { Comment } from "@/lib/generated/issue-tracker";

export const commentFixtures: Comment[] = [
  {
    id: 1,
    issueId: 1,
    author: "admin",
    type: "progress",
    body: "Mapped the frontend API calls that need MSW coverage.",
    createdAt: "2026-06-01T01:10:00.000Z",
  },
  {
    id: 2,
    issueId: 3,
    author: "qa",
    type: "general",
    body: "Use the board select controls to verify status changes are reflected in the summary.",
    createdAt: "2026-06-01T03:15:00.000Z",
  },
  {
    id: 3,
    issueId: 4,
    author: "reviewer",
    type: "handoff",
    body: "Detail page comments are loaded through the generated issue-tracker client.",
    createdAt: "2026-06-01T04:20:00.000Z",
  },
  {
    id: 4,
    issueId: 9,
    author: "agent",
    type: "blocker",
    body: "GitHub credentials are not configured in the development environment.",
    createdAt: "2026-06-01T09:15:00.000Z",
  },
];
