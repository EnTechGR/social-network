/**
 * app/about/page.tsx
 *
 * About page - accessible at /about
 * Each folder with a page.tsx file becomes a route in Next.js
 */

import { Metadata } from 'next';

// Metadata for SEO
export const metadata: Metadata = {
  title: 'About - Social Network',
  description: 'Learn more about our social network',
};

export default function AboutPage() {
  return (
    <div className="max-w-4xl mx-auto px-4 py-12">
      <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-6">
        About Us
      </h1>

      <div className="prose dark:prose-invert">
        <p className="text-lg text-gray-600 dark:text-gray-300">
          Welcome to our social network! This is a placeholder page to demonstrate
          how routing works in Next.js.
        </p>

        <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mt-8 mb-4">
          Our Mission
        </h2>
        <p className="text-gray-600 dark:text-gray-300">
          To connect people from around the world and build meaningful relationships.
        </p>

        <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mt-8 mb-4">
          Key Features
        </h2>
        <ul className="list-disc list-inside text-gray-600 dark:text-gray-300 space-y-2">
          <li>Connect with friends and family</li>
          <li>Share moments and updates</li>
          <li>Discover new content</li>
          <li>Join communities</li>
        </ul>
      </div>
    </div>
  );
}
