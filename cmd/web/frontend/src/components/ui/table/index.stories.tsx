import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { Pagination } from "@/components/ui/pagination";
import { IssueTable } from "./index";

const meta = {
  title: "UI/Table",
  component: IssueTable,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    issues: issueFixtures.slice(0, 6),
    onSortChange: () => undefined,
    sortBy: "updated_at",
    sortDirection: "desc",
  },
} satisfies Meta<typeof IssueTable>;

export default meta;

type Story = StoryObj<typeof meta>;

export const IssueRows: Story = {};

export const SortedByID: Story = {
  args: {
    sortBy: "id",
    sortDirection: "asc",
  },
};

export const WithPagination: Story = {
  render: (args) => (
    <div style={{ display: "grid", gap: 14 }}>
      <IssueTable {...args} />
      <Pagination
        nextDisabled={false}
        nextLabel="Next"
        onNext={() => undefined}
        onPrevious={() => undefined}
        previousDisabled
        previousLabel="Previous"
        summary="1-50 of 128 issues"
      />
    </div>
  ),
};
