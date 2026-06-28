import type { Meta, StoryObj } from "@storybook/react-vite";
import { MarkdownEditor } from "./index";

const labels = {
  cancel: "Cancel",
  edit: "Edit description",
  empty: "No description",
  raw: "Raw",
  preview: "Preview",
  save: "Save",
  saving: "Saving...",
  textarea: "Description",
};

const content = `## Goal

Edit the issue description with **Markdown** support.

- Keep the read view quiet
- Preview tables, task lists, and code blocks
- Preserve user input while editing
`;

const meta = {
  title: "UI/MarkdownEditor",
  component: MarkdownEditor,
  args: {
    labels,
    title: "Description",
    titleID: "markdown-editor-story",
    value: content,
  },
} satisfies Meta<typeof MarkdownEditor>;

export default meta;

type Story = StoryObj<typeof meta>;

export const ReadOnly: Story = {};

export const EditingRaw: Story = {
  args: {
    initialMode: "edit",
    initialTab: "raw",
  },
};

export const EditingPreview: Story = {
  args: {
    initialMode: "edit",
    initialTab: "preview",
  },
};

export const InitialEditMode: Story = {
  args: {
    initialMode: "edit",
    initialTab: "raw",
    showActions: false,
    title: undefined,
  },
};
