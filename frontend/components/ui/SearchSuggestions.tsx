'use client';

import React, { useState, useEffect } from "react";
import Link from "next/link";
import Tabs from "./Tabs";
import { searchUsers } from "@/lib/api";

interface SearchSuggestionsProps {
	searchQuery?: string;
	onSelectUser?: () => void;
}

function UserSuggestion({ user, onSelect }: { user: any; onSelect?: () => void }) {
	const fullName = `${user.first_name} ${user.last_name}`.trim() || user.nickname || 'User';
	const subtitle = user.nickname ? `@${user.nickname}` : user.email;
	
	return (
		<Link href={`/profile/${user.id}`} className="no-underline" onClick={onSelect}>
			<div
				className="
					flex
					w-full
					min-h-[80px]
					p-4
					items-center
					gap-3
					rounded-none
					border-b
					border-parea-border
					bg-parea-white
					hover:bg-parea-yellow
					cursor-pointer
					transition-colors
					duration-200
				"
			>
				{/* Avatar */}
				<div className="w-12 h-12 rounded-full bg-parea-black text-parea-white flex items-center justify-center font-semibold shrink-0">
					{user.first_name?.[0]?.toUpperCase() || user.nickname?.[0]?.toUpperCase() || 'U'}
				</div>
				
				{/* User Info */}
				<div className="flex flex-col items-start gap-1 flex-1">
					<div className="text-parea-black font-sans text-lg font-semibold leading-[1.4]">
						{fullName}
					</div>
					<div className="text-parea-black/60 font-sans text-sm font-normal leading-[1.5]">
						{subtitle}
					</div>
				</div>
				
				{/* Online Status */}
				{user.is_online && (
					<div className="w-3 h-3 rounded-full bg-green-500" title="Online" />
				)}
			</div>
		</Link>
	);
}

export default function SearchSuggestions({ searchQuery = '', onSelectUser }: SearchSuggestionsProps) {
	const [activeTab, setActiveTab] = useState<string>('Users');
	const [users, setUsers] = useState<any[]>([]);
	const [loading, setLoading] = useState(false);

	useEffect(() => {
		if (activeTab === 'Users') {
			loadUsers();
		}
	}, [searchQuery, activeTab]);

	async function loadUsers() {
		setLoading(true);
		try {
			const result = await searchUsers(searchQuery);
			setUsers(result.users || []);
		} catch (error) {
			console.error('Failed to search users:', error);
			setUsers([]);
		} finally {
			setLoading(false);
		}
	}

	return (
		<div
			className="
				flex
				flex-col
				w-full
				max-w-full
				items-center
				border-b
				border-parea-black
				bg-parea-white
			"
		>
			<div
				className="
					flex
					flex-col
					min-h-[150px]
					max-h-[500px]
					px-8
					pt-4
					pb-8
					gap-4
					items-start
					self-stretch
					bg-parea-white
					rounded-xl
					shadow-lg
					absolute
					top-full
					left-0
					z-10
					w-full
				"
			>
				<div>
					<Tabs 
						tabs={["Users", "Posts", "Events", "Groups"]} 
						defaultTab="Users"
						onTabChange={setActiveTab}
					/>
				</div>
				<div className="flex flex-col gap-0 w-full overflow-y-auto">
					{activeTab === 'Users' && (
						<>
							{loading ? (
								<p className="text-parea-black/60 p-4">Searching...</p>
							) : users.length === 0 ? (
								<p className="text-parea-black/60 p-4">
									{searchQuery ? 'No users found' : 'Start typing to search for users'}
								</p>
							) : (
								users.map((user) => <UserSuggestion key={user.id} user={user} onSelect={onSelectUser} />)
							)}
						</>
					)}
					{activeTab === 'Posts' && (
						<p className="text-parea-black/60 p-4">Post search coming soon...</p>
					)}
					{activeTab === 'Events' && (
						<p className="text-parea-black/60 p-4">Event search coming soon...</p>
					)}
					{activeTab === 'Groups' && (
						<p className="text-parea-black/60 p-4">Group search coming soon...</p>
					)}
				</div>
			</div>
		</div>
	);
}
