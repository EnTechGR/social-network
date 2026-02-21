'use client';

import { Search } from 'lucide-react';
import SearchSuggestions from './SearchSuggestions';
import { useEffect, useRef, useState } from 'react';
import Button from './Button';
import Link from 'next/link';
import CreatePostModal from './CreatePostModal';
import { getForumUsers } from '@/lib/api';

// --- SearchInputWithDropdown component ---
function SearchInputWithDropdown() {
  const [showDropdown, setShowDropdown] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setShowDropdown(false);
      }
    }

    document.addEventListener('mousedown', handleClick);

    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  return (
    <div ref={containerRef} className="relative w-full">
      <input
        type="text"
        placeholder="Search..."
        className="
          flex-1
          border-none
          outline-none
          bg-transparent
          text-black/60
          font-mono
          text-[15px]
          font-medium
          leading-[1.5]
          tracking-[-0.15px]
          uppercase
          placeholder:text-black/60
          w-full
        "
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        onFocus={() => setShowDropdown(true)}
      />
      {showDropdown && (
        <div
          className="
            absolute
            left-0
            top-[calc(100%+14px)]
            z-50
            w-[620px]
            max-w-[90vw]
          "
        >
          <SearchSuggestions
            query={searchQuery}
            onClose={() => setShowDropdown(false)}
          />
        </div>
      )}
    </div>
  );
}

export const Navbar = () => {
  const [isCreatePostOpen, setIsCreatePostOpen] = useState(false);
  const [selectableUsers, setSelectableUsers] = useState<Array<{ id: string; name: string; avatarUrl?: string }>>([]);

  useEffect(() => {
    let cancelled = false;

    getForumUsers()
      .then((users) => {
        if (cancelled) return;
        const mapped = users.map((user) => {
          const fullName = `${user.first_name || ''} ${user.last_name || ''}`.trim();
          return {
            id: user.id,
            name: fullName || user.nickname || user.email || user.id,
            avatarUrl: undefined,
          };
        });
        setSelectableUsers(mapped);
      })
      .catch((err) => {
        if (!cancelled) {
          console.warn('Failed to load selectable users for private posts:', err);
          setSelectableUsers([]);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <>
    <nav
      className="
        w-full
        border-b
        border-parea-black
        bg-parea-white
        flex
        justify-center
        items-center
        p-0
      "
    >
      <div
        className="
          w-full
          max-w-[1349px]
          h-[72px]
          flex
          flex-row
          items-center
          justify-between
          px-16
          box-border
        "
      >
        {/* Logo */}
        <div className="
          flex
          w-[115px]
          justify-center
          items-center
          gap-[5.61px]
        ">
          <img src="/logo.svg" alt="Logo" className="
            h-8
            w-auto
          " />
        </div>

        {/* Search */}
        <div
          className="
            relative
            hidden
            md:flex
            w-[299px]
            h-[47px]
            px-3
            py-2
            items-center
            gap-3
            rounded-full
            border
            border-parea-border
            bg-parea-white
            ml-8
          "
        >
          <div className="
            w-6
            h-6
            flex-shrink-0
            flex
            items-center
            justify-center
          ">
            <Search className="w-6 h-6 text-parea-black/70" />
          </div>
          {/* Search input with focus/blur handlers */}
          <SearchInputWithDropdown />
        </div>

        {/* Links and Button */}
        <div className="
          flex
          items-center
          gap-[30px]
          ml-auto
        ">
          <Link href="/feed" className="no-underline">
            <div className="
              flex
              justify-center
              items-center
              gap-1
              cursor-pointer
            ">
              <span className="
                text-parea-black
                font-mono
                text-[15px]
                font-medium
                leading-[1.5]
                tracking-[-0.15px]
                uppercase
              ">
                Feed
              </span>
            </div>
          </Link>

          <Link href="/profile" className="no-underline">
            <div className="
              flex
              justify-center
              items-center
              gap-1
              cursor-pointer
            ">
              <span className="
                text-parea-black
                font-mono
                text-[15px]
                font-medium
                leading-[1.5]
                tracking-[-0.15px]
                uppercase
              ">
                Profile
              </span>
            </div>
          </Link>

          <Button 
            variant="primary" 
            size="lg" 
            className="cursor-pointer"
            onClick={() => setIsCreatePostOpen(true)}
          >
            Create Post
          </Button>
        </div>
      </div>
    </nav>

    <CreatePostModal 
      isOpen={isCreatePostOpen}
      onClose={() => setIsCreatePostOpen(false)}
      followers={selectableUsers}
    />
    </>
  );
};

export default Navbar;
