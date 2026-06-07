import type { Preview } from "@storybook/react-vite";
import { MemoryRouter } from "react-router-dom";
import "../src/app/globals.css";
import "../src/lib/i18n";

const preview: Preview = {
  decorators: [
    (Story) => (
      <MemoryRouter initialEntries={["/issues"]}>
        <Story />
      </MemoryRouter>
    ),
  ],
  parameters: {
    controls: {
      expanded: true,
    },
  },
};

export default preview;
