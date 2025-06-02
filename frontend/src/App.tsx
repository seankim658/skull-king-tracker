import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import "./App.css";
import { Toaster } from "./components/ui/sonner";
import { ProtectedRoute } from "./auth/protected-route";

// Providers
import { AuthProvider } from "./providers/auth-provider";
import { ThemeProvider } from "./providers/theme-provider";

// Hooks
import { useAuth } from "./hooks/use-auth";

// Layouts
import { MainLayout } from "./layouts/main-layout";

// Pages
import { LoginPage } from "./pages/login-page";
import { DashboardPage } from "./pages/dashboard-page";
import { SettingsPage } from "./pages/settings-page";
import { StartSessionPage } from "./pages/start-session-page";
import { ProfilePage } from "./pages/profile-page";

function AppContent() {
  const { isLoadingAuth } = useAuth();

  if (isLoadingAuth) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p>Initializing Skull King App...</p>
      </div>
    );
  }

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />

      {/* Protected Routes */}
      <Route element={<ProtectedRoute />}>
        <Route element={<MainLayout />}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/start-session" element={<StartSessionPage />} />
          <Route path="/users/:userId" element={<ProfilePage />} />
        </Route>
      </Route>

      {/* Catch-all route, should ideally redirect to a 404 page or home */}
      {/* For now, redirecting to login if no match, or dashboard if authenticated */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

function App() {
  return (
    <Router>
      <AuthProvider>
        <ThemeProvider>
          <AppContent />
          <Toaster richColors closeButton />
        </ThemeProvider>
      </AuthProvider>
    </Router>
  );
}

export default App;
