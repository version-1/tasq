import { Navigate, Route, Routes } from "react-router-dom";
import DashboardPage from "./app/dashboard/page";
import { IssueDetailLayout } from "./app/issues/[id]/_components/issue-detail-layout";
import ConversationRoute from "./app/issues/[id]/conversations/page";
import IssueDetailRoute from "./app/issues/[id]/page";
import IssuesPage from "./app/issues/page";
import SettingsPage from "./app/settings/page";
import { DefaultLayout } from "./components/layout/default";
import { Layout } from "./components/layout";
import { ProjectLayout } from "./components/layout/project";

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/issues" replace />} />
        <Route
          path="/issues"
          element={<ProjectLayout><IssuesPage /></ProjectLayout>}
        />
        <Route
          path="/projects/:projectKey/issues"
          element={<ProjectLayout><IssuesPage /></ProjectLayout>}
        />
        <Route
          path="/issues/:id"
          element={<IssueDetailLayout><IssueDetailRoute /></IssueDetailLayout>}
        />
        <Route
          path="/issues/:id/conversations"
          element={<IssueDetailLayout><ConversationRoute /></IssueDetailLayout>}
        />
        <Route
          path="/issues/:id/runs/:runId/conversations"
          element={<IssueDetailLayout><ConversationRoute /></IssueDetailLayout>}
        />
        <Route
          path="/dashboard"
          element={<DefaultLayout><DashboardPage /></DefaultLayout>}
        />
        <Route
          path="/settings"
          element={<ProjectLayout><SettingsPage /></ProjectLayout>}
        />
        <Route path="*" element={<Navigate to="/issues" replace />} />
      </Routes>
    </Layout>
  );
}
