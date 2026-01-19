'use client';

import React from 'react';
import { Search } from 'lucide-react';
import Button from './Button';
import Link from 'next/link';

export const Navbar = () => {
  return (
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
            "
          />
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

          <Button variant="primary" size="lg" className="cursor-pointer">
            Create Post
          </Button>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
