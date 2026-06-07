import { Navigate, Route, Routes } from "react-router-dom";
import AgentsPage from "./app/agents/page";
import DashboardPage from "./app/dashboard/page";
import IssueDetailRoute from "./app/issues/[id]/page";
import IssuesPage from "./app/issues/page";
import SettingsPage from "./app/settings/page";
import { DashboardLayout, Layout, ProjectsLayout } from "./components/layout";

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/issues" replace />} />
        <Route
          path="/issues"
          element={<ProjectsLayout><IssuesPage /></ProjectsLayout>}
        />
        <Route
          path="/projects/:projectKey/issues"
          element={<ProjectsLayout><IssuesPage /></ProjectsLayout>}
        />
        <Route
          path="/issues/:id"
          element={<ProjectsLayout><IssueDetailRoute /></ProjectsLayout>}
        />
        <Route
          path="/agents"
          element={<ProjectsLayout><AgentsPage /></ProjectsLayout>}
        />
        <Route
          path="/dashboard"
          element={<DashboardLayout><DashboardPage /></DashboardLayout>}
        />
        <Route
          path="/settings"
          element={<ProjectsLayout><SettingsPage /></ProjectsLayout>}
        />
        <Route path="*" element={<Navigate to="/issues" replace />} />
      </Routes>
    </Layout>
  );
}
