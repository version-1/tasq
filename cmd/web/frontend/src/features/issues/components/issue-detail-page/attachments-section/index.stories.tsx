import type { Meta, StoryObj } from "@storybook/react-vite";
import type { Attachment } from "@/lib/types";
import { AttachmentsSection } from "./index";

const attachments: Attachment[] = [
  {
    id: "attachment-1",
    entityType: "issue",
    entityId: "42",
    filename: "issue-detail-reference.png",
    contentType: "image/png",
    size: 124_928,
    createdAt: "2026-06-20T01:00:00.000Z",
  },
  {
    id: "attachment-2",
    entityType: "issue",
    entityId: "42",
    filename: "flow-diagram.webp",
    contentType: "image/webp",
    size: 934_400,
    createdAt: "2026-06-20T01:30:00.000Z",
  },
];

const meta = {
  title: "Features/Issues/IssueDetail/AttachmentsSection",
  component: AttachmentsSection,
  args: {
    attachments,
    error: "",
    isLoading: false,
  },
} satisfies Meta<typeof AttachmentsSection>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Populated: Story = {};

export const Empty: Story = {
  args: {
    attachments: [],
  },
};
