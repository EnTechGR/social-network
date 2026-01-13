/**
 * app/dashboard/settings/page.tsx
 *
 * Settings page - accessible at /dashboard/settings
 * Nested route: dashboard > settings
 */

'use client'; // This is a Client Component because it uses state

import { useState } from 'react';
import Card from '@/components/ui/Card';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';

export default function SettingsPage() {
  // State management example
  const [name, setName] = useState('Jane Doe');
  const [email, setEmail] = useState('jane@example.com');
  const [bio, setBio] = useState('Software developer');

  const handleSave = () => {
    // In a real app, you'd send this data to an API
    console.log('Saving settings:', { name, email, bio });
    alert('Settings saved!');
  };

  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-8">
        Account Settings
      </h1>

      <div className="max-w-2xl">
        <Card>
          <Card.Header>
            <Card.Title>Profile Information</Card.Title>
          </Card.Header>

          <Card.Content>
            <div className="space-y-4">
              <Input
                label="Full Name"
                value={name}
                onChange={setName}
                placeholder="Enter your name"
                required
              />

              <Input
                label="Email"
                type="email"
                value={email}
                onChange={setEmail}
                placeholder="Enter your email"
                required
              />

              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  Bio
                </label>
                <textarea
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  rows={4}
                  className="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Tell us about yourself"
                />
              </div>

              <div className="flex gap-3 pt-4">
                <Button variant="primary" onClick={handleSave}>
                  Save Changes
                </Button>
                <Button variant="secondary" onClick={() => {
                  setName('Jane Doe');
                  setEmail('jane@example.com');
                  setBio('Software developer');
                }}>
                  Reset
                </Button>
              </div>
            </div>
          </Card.Content>
        </Card>
      </div>
    </div>
  );
}
