import { Navigate, Route, Routes } from "react-router-dom";
import DashboardPage from "./app/dashboard/page";
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
          element={<ProjectLayout><IssueDetailRoute /></ProjectLayout>}
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
