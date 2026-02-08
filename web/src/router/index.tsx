import { Navigate, createBrowserRouter } from 'react-router-dom';
import MainLayout from '../layouts/MainLayout';
import AdminLayout from '../layouts/AdminLayout';
import Dashboard from '../pages/Dashboard';
import Challenges from '../pages/Challenges';
import Leaderboard from '../pages/Leaderboard';
import Profile from '../pages/Profile';
import Login from '../pages/Login';
import AdminLogin from '../pages/Admin/Login';
import AdminOverview from '../pages/Admin/Overview';
import AdminChallenges from '../pages/Admin/Challenges';
import AdminChallengeNew from '../pages/Admin/ChallengeNew';
import AdminInstances from '../pages/Admin/Instances';
import AdminSubmissions from '../pages/Admin/Submissions';
import AdminImages from '../pages/Admin/Images';
import AdminLogs from '../pages/Admin/Logs';

const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/admin/login',
    element: <AdminLogin />,
  },
  {
    path: '/admin',
    element: <AdminLayout />,
    children: [
      {
        path: '/admin',
        element: <Navigate to="/admin/overview" replace />,
      },
      {
        path: '/admin/overview',
        element: <AdminOverview />,
      },
      {
        path: '/admin/challenges',
        element: <AdminChallenges />,
      },
      {
        path: '/admin/challenges/new',
        element: <AdminChallengeNew />,
      },
      {
        path: '/admin/instances',
        element: <AdminInstances />,
      },
      {
        path: '/admin/submissions',
        element: <AdminSubmissions />,
      },
      {
        path: '/admin/images',
        element: <AdminImages />,
      },
      {
        path: '/admin/logs',
        element: <AdminLogs />,
      },
    ],
  },
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        path: '/',
        element: <Dashboard />,
      },
      {
        path: '/challenges',
        element: <Challenges />,
      },
      {
        path: '/leaderboard',
        element: <Leaderboard />,
      },
      {
        path: '/profile',
        element: <Profile />,
      },
    ],
  },
]);

export default router;
