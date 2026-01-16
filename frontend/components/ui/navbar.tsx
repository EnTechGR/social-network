'use client';

import React from 'react';
import Button from './Button';
import Link from 'next/link';

export const Navbar = () => {
  return (
    <nav
      style={{
        width: '100%',
        borderBottom: '1px solid var(--parea-black)',
        background: 'var(--parea-white)',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        padding: 0,
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: '1349px',
          height: '72px',
          display: 'flex',
          flexDirection: 'row',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 64px',
          boxSizing: 'border-box',
        }}
      >
        {/* Logo */}
        <div
          style={{
            display: 'flex',
            width: '115px',
            justifyContent: 'center',
            alignItems: 'center',
            gap: '5.61px',
          }}
        >
          <img src="/logo.svg" alt="Logo" style={{ height: '32px', width: 'auto' }} />
        </div>

        {/* Search */}
        <div
          style={{
            display: 'flex',
            width: '299px',
            height: '47px',
            padding: '8px 12px',
            alignItems: 'center',
            gap: '12px',
            borderRadius: '50px',
            border: '1px solid var(--parea-border)',
            background: 'var(--parea-white)',
            marginLeft: '32px',
          }}
        >
          <div
            style={{
              width: '24px',
              height: '24px',
              flexShrink: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {/* ...existing SVG... */}
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M9.41503 15.8765C7.58436 15.8765 6.0347 15.2415 4.76603 13.9715C3.49753 12.7015 2.86328 11.1693 2.86328 9.3748C2.86328 7.5803 3.49828 6.04805 4.76828 4.77805C6.03811 3.50805 7.57453 2.87305 9.37753 2.87305C11.1804 2.87305 12.7125 3.50805 13.974 4.77805C15.2357 6.04805 15.8665 7.58155 15.8665 9.37855C15.8665 10.1007 15.7529 10.7921 15.5255 11.4528C15.2982 12.1135 14.9572 12.7335 14.5025 13.3128L20.5165 19.277C20.6865 19.4435 20.7715 19.6499 20.7715 19.896C20.7715 20.1424 20.6865 20.3505 20.5165 20.5205C20.3445 20.6944 20.1349 20.7813 19.8875 20.7813C19.6404 20.7813 19.4355 20.6944 19.273 20.5205L13.284 14.5373C12.788 14.9583 12.208 15.2867 11.544 15.5225C10.88 15.7585 10.1704 15.8765 9.41503 15.8765ZM9.39203 14.173C10.7209 14.173 11.8483 13.7044 12.7743 12.767C13.7003 11.8297 14.1633 10.699 14.1633 9.3748C14.1633 8.05063 13.6999 6.91988 12.7733 5.98255C11.8468 5.04521 10.7197 4.57655 9.39203 4.57655C8.05053 4.57655 6.91086 5.04521 5.97303 5.98255C5.03536 6.91988 4.56653 8.05063 4.56653 9.3748C4.56653 10.699 5.03511 11.8297 5.97228 12.767C6.90945 13.7044 8.04936 14.173 9.39203 14.173Z" fill="#121214" fillOpacity="0.7"/>
              <path d="M9.37793 3.37305C11.0485 3.37314 12.4523 3.95621 13.6191 5.13086C14.7873 6.30676 15.3662 7.71344 15.3662 9.37891C15.3662 10.0474 15.2614 10.6836 15.0527 11.29C14.8445 11.8951 14.5314 12.4661 14.1094 13.0039L13.834 13.3545L14.1504 13.668L20.1641 19.6318L20.167 19.6338C20.2373 19.7027 20.2715 19.7795 20.2715 19.8965C20.2714 19.9844 20.2519 20.0513 20.2109 20.1104L20.1631 20.167L20.1611 20.1689C20.0843 20.2466 20.0031 20.2812 19.8877 20.2812C19.7729 20.2812 19.7026 20.2471 19.6387 20.1787L19.6328 20.1729L19.626 20.167L13.6377 14.1836L13.3115 13.8584L12.9609 14.1562C12.5707 14.4875 12.1189 14.7578 11.6025 14.9658L11.377 15.0518C10.7705 15.2673 10.1174 15.377 9.41504 15.377C7.7142 15.377 6.29291 14.7921 5.12012 13.6182C3.94528 12.4419 3.36333 11.0369 3.36328 9.375C3.36328 7.71324 3.94575 6.30816 5.12207 5.13184C6.29772 3.95603 7.70696 3.37305 9.37793 3.37305ZM9.3916 4.07617C8.01159 4.07627 6.8119 4.53393 5.81641 5.44141L5.61914 5.62891C4.58792 6.65985 4.06641 7.9181 4.06641 9.375C4.06645 10.8318 4.58838 12.0892 5.61914 13.1201C6.65134 14.1525 7.91843 14.6728 9.3916 14.6729C10.8528 14.6729 12.109 14.1515 13.1299 13.1182C14.1487 12.0868 14.663 10.8299 14.6631 9.375C14.6631 7.9199 14.1486 6.66228 13.1289 5.63086C12.1077 4.59777 10.8517 4.07617 9.3916 4.07617Z" stroke="#121214" strokeOpacity="0.7"/>
            </svg>
          </div>
          <input
            type="text"
            placeholder="Search..."
            style={{
              flex: 1,
              border: 'none',
              outline: 'none',
              background: 'transparent',
              color: 'rgba(0, 0, 0, 0.60)',
              fontFamily: '"IBM Plex Mono"',
              fontSize: '15px',
              fontStyle: 'normal',
              fontWeight: 500,
              lineHeight: '150%',
              letterSpacing: '-0.15px',
              textTransform: 'uppercase',
            }}
          />
        </div>

        {/* Links and Button */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '30px',
            marginLeft: 'auto',
          }}
        >
          <Link href="/feed" style={{ textDecoration: 'none' }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                gap: '4px',
                cursor: 'pointer',
              }}
            >
              <span
                style={{
                  color: 'var(--parea-black)',
                  fontFamily: '"IBM Plex Mono"',
                  fontSize: '15px',
                  fontStyle: 'normal',
                  fontWeight: 500,
                  lineHeight: '150%',
                  letterSpacing: '-0.15px',
                  textTransform: 'uppercase',
                }}
              >
                Feed
              </span>
            </div>
          </Link>

          <Link href="/profile" style={{ textDecoration: 'none' }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                gap: '4px',
                cursor: 'pointer',
              }}
            >
              <span
                style={{
                  color: 'var(--parea-black)',
                  fontFamily: '"IBM Plex Mono"',
                  fontSize: '15px',
                  fontStyle: 'normal',
                  fontWeight: 500,
                  lineHeight: '150%',
                  letterSpacing: '-0.15px',
                  textTransform: 'uppercase',
                }}
              >
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
