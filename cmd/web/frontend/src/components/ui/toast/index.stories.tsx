import { useEffect } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { toastStore, type ToastType } from "@/lib/toast";
import { ToastStack } from "./index";

type ToastStackStoryProps = {
  initialToasts: ToastType[];
};

function ToastStackStory({ initialToasts }: ToastStackStoryProps) {
  useEffect(() => {
    toastStore.clear();
    initialToasts.forEach((type) => {
      toastStore[type]({ message: toastMessage(type) });
    });

    return () => {
      toastStore.clear();
    };
  }, [initialToasts]);

  return (
    <div style={{ minHeight: 220, padding: 24 }}>
      <div style={{ display: "flex", gap: 8 }}>
        <button
          type="button"
          onClick={() => toastStore.success({ message: toastMessage("success") })}
        >
          Show success
        </button>
        <button
          type="button"
          onClick={() => toastStore.error({ message: toastMessage("error") })}
        >
          Show error
        </button>
      </div>
      <ToastStack />
    </div>
  );
}

const meta = {
  title: "UI/Toast",
  component: ToastStack,
  render: (args) => <ToastStackStory {...args} />,
  args: {
    initialToasts: ["success", "error"],
  },
} satisfies Meta<typeof ToastStackStory>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Stack: Story = {};

export const Success: Story = {
  args: {
    initialToasts: ["success"],
  },
};

export const Error: Story = {
  args: {
    initialToasts: ["error"],
  },
};

function toastMessage(type: ToastType): string {
  if (type === "error") {
    return "Could not update the issue status. Please try again.";
  }
  return "Issue status updated.";
}
