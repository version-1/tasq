import type { Meta, StoryObj } from "@storybook/react-vite";
import type { Project } from "@/lib/types";
import { Sidebar } from "./index";

const projects: Project[] = [
  {
    id: 1,
    key: "tasq",
    name: "Tasq",
    description: "Local issue tracking workspace",
    location: "/workspace/tasq",
    createdAt: "2026-06-01T00:00:00Z",
    updatedAt: "2026-06-07T00:00:00Z",
  },
  {
    id: 2,
    key: "agent-docs",
    name: "Agent-docs",
    description: "Agent documentation workspace",
    location: "/workspace/agent-docs",
    createdAt: "2026-06-01T00:00:00Z",
    updatedAt: "2026-06-07T00:00:00Z",
  },
  {
    id: 3,
    key: "some-project",
    name: "Some Project",
    description: "Sample project workspace",
    location: "/workspace/some-project",
    createdAt: "2026-06-01T00:00:00Z",
    updatedAt: "2026-06-07T00:00:00Z",
  },
];

const meta = {
  title: "Layout/Sidebar",
  component: Sidebar,
  args: {
    activePage: "issues",
    activeProjectID: 1,
    onAddProject: () => undefined,
    projects,
  },
} satisfies Meta<typeof Sidebar>;

export default meta;

type Story = StoryObj<typeof meta>;

export const ActiveProject: Story = {};

export const AllProjects: Story = {
  args: {
    activePage: "issues",
    activeProjectID: null,
  },
};

export const Dashboard: Story = {
  args: {
    activePage: "dashboard",
    activeProjectID: null,
  },
};
