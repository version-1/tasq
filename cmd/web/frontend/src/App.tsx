import { Navigate, Route, Routes } from "react-router-dom";
import DashboardPage from "./app/dashboard/page";
import { IssueDetailLayout } from "./components/layout/issue";
import ConversationRoute from "./app/issues/[id]/conversations/page";
import IssueDetailRoute from "./app/issues/[id]/page";
import IssuesPage from "./app/issues/page";
import SettingsPage from "./app/settings/page";
import { DefaultLayout } from "./components/layout/default";
import type { HeaderPageLink } from "./components/layout/header";
import { Layout } from "./components/layout";
import { ProjectLayout } from "./components/layout/project";

const headerPages = [
  { key: "issues", href: "/issues", titleKey: "header.board" },
  { key: "settings", href: "/settings", titleKey: "header.settings" },
] satisfies readonly HeaderPageLink[];

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/issues" replace />} />
        <Route
          path="/issues"
          element={<ProjectLayout pages={headerPages}><IssuesPage /></ProjectLayout>}
        />
        <Route
          path="/projects/:projectKey/issues"
          element={<ProjectLayout pages={headerPages}><IssuesPage /></ProjectLayout>}
        />
        <Route
          path="/issues/:id"
          element={<IssueDetailLayout pages={headerPages}><IssueDetailRoute /></IssueDetailLayout>}
        />
        <Route
          path="/issues/:id/conversations"
          element={<IssueDetailLayout pages={headerPages}><ConversationRoute /></IssueDetailLayout>}
        />
        <Route
          path="/issues/:id/runs/:runId/conversations"
          element={<IssueDetailLayout pages={headerPages}><ConversationRoute /></IssueDetailLayout>}
        />
        <Route
          path="/dashboard"
          element={<DefaultLayout pages={headerPages}><DashboardPage /></DefaultLayout>}
        />
        <Route
          path="/settings"
          element={<ProjectLayout pages={headerPages}><SettingsPage /></ProjectLayout>}
        />
        <Route path="*" element={<Navigate to="/issues" replace />} />
      </Routes>
    </Layout>
  );
}
