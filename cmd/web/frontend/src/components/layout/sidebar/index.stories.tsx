import type { Meta, StoryObj } from "@storybook/react-vite";
import type { Project } from "@/lib/types";
import { Sidebar } from "./index";

const projects: Project[] = [
  {
    id: 1,
    key: "product-website",
    name: "Product Website",
    description: "Public website workstream",
    location: "/workspace/product-website",
    createdAt: "2026-06-01T00:00:00Z",
    updatedAt: "2026-06-07T00:00:00Z",
  },
  {
    id: 2,
    key: "api-backend",
    name: "API Backend",
    description: "Issue tracker API",
    location: "/workspace/api-backend",
    createdAt: "2026-06-01T00:00:00Z",
    updatedAt: "2026-06-07T00:00:00Z",
  },
  {
    id: 3,
    key: "mobile-app",
    name: "Mobile App",
    description: "Mobile client",
    location: "/workspace/mobile-app",
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
