/**
 * app/dashboard/page.tsx
 *
 * Dashboard overview page - accessible at /dashboard
 * This page is wrapped by both the root layout AND the dashboard layout
 */

import Card from '@/components/ui/Card';

export default function DashboardPage() {
  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-8">
        Dashboard Overview
      </h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {/* Stats Cards */}
        <Card>
          <Card.Title>Total Posts</Card.Title>
          <Card.Content>
            <p className="text-4xl font-bold text-blue-600 mt-2">127</p>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              +12 this week
            </p>
          </Card.Content>
        </Card>

        <Card>
          <Card.Title>Followers</Card.Title>
          <Card.Content>
            <p className="text-4xl font-bold text-green-600 mt-2">1,234</p>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              +45 this week
            </p>
          </Card.Content>
        </Card>

        <Card>
          <Card.Title>Engagement</Card.Title>
          <Card.Content>
            <p className="text-4xl font-bold text-purple-600 mt-2">89%</p>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              +3% this week
            </p>
          </Card.Content>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <Card.Header>
          <Card.Title>Recent Activity</Card.Title>
        </Card.Header>
        <Card.Content>
          <div className="space-y-4">
            <div className="flex items-center gap-4">
              <div className="w-2 h-2 bg-blue-600 rounded-full"></div>
              <p className="text-gray-600 dark:text-gray-300">
                New follower: @johndoe
              </p>
              <span className="text-sm text-gray-400 ml-auto">2h ago</span>
            </div>
            <div className="flex items-center gap-4">
              <div className="w-2 h-2 bg-green-600 rounded-full"></div>
              <p className="text-gray-600 dark:text-gray-300">
                Your post got 50 likes
              </p>
              <span className="text-sm text-gray-400 ml-auto">5h ago</span>
            </div>
            <div className="flex items-center gap-4">
              <div className="w-2 h-2 bg-purple-600 rounded-full"></div>
              <p className="text-gray-600 dark:text-gray-300">
                @janedoe commented on your post
              </p>
              <span className="text-sm text-gray-400 ml-auto">1d ago</span>
            </div>
          </div>
        </Card.Content>
      </Card>
    </div>
  );
}
