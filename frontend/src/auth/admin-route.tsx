import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";

export function AdminRoute() {
  const { user } = useAuth();

  if (user?.role !== "superuser") {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}
