# Parea Social Network

A modern social network web application built with Next.js, React, and Tailwind CSS.

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

---

## 🎨 Design System

This project uses the **Parea Design System** with custom colors, typography, and components.

### Colors

| Variable | Hex | Usage |
|----------|-----|-------|
| `parea-black` | #121214 | Text, borders |
| `parea-white` | #F0EFEF | Backgrounds |
| `parea-grey` | #D7D7D7 | Muted elements |
| `parea-yellow` | #DDFF30 | Primary accent |
| `parea-border` | rgba(18,18,20,0.7) | Borders |

### Typography

- **Headings**: Inter (h1-h6)
- **Labels**: IBM Plex Mono (buttons, tabs, special text)

---

## 🧩 Components

### Button

Text buttons with multiple variants and sizes.

```tsx
import Button from '@/components/ui/Button';

// Primary buttons (yellow background)
<Button variant="primary" size="lg">Sign Up</Button>
<Button variant="primary" size="sm">Submit</Button>

// Secondary buttons (transparent background)
<Button variant="secondary" size="lg">Learn More</Button>
<Button variant="secondary" size="sm">Cancel</Button>

// Tertiary buttons (underline style, for tabs)
<Button variant="tertiary">Posts</Button>
<Button variant="tertiary" isActive>Comments</Button>

// With icons
<Button variant="primary" leftIcon={<PlusIcon />}>Add Friend</Button>
<Button variant="primary" rightIcon={<ArrowIcon />}>Continue</Button>

// Loading state
<Button variant="primary" isLoading>Saving...</Button>

// Disabled
<Button variant="primary" disabled>Disabled</Button>
```

#### Button Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `variant` | `'primary' \| 'secondary' \| 'tertiary'` | `'primary'` | Button style |
| `size` | `'sm' \| 'lg'` | `'lg'` | Button size |
| `isActive` | `boolean` | `false` | Active state (tertiary only) |
| `leftIcon` | `ReactNode` | - | Icon before text |
| `rightIcon` | `ReactNode` | - | Icon after text |
| `isLoading` | `boolean` | `false` | Show loading spinner |
| `disabled` | `boolean` | `false` | Disable button |

---

### IconButton

Icon-only buttons with hover animations.

```tsx
import IconButton from '@/components/ui/IconButton';

// Arrow button (rotates -45° on hover)
<IconButton variant="arrow" aria-label="Next" />
<IconButton variant="arrow" aria-label="Go to profile" onClick={handleClick} />

// Close button (rotates 180° on hover)
<IconButton variant="close" aria-label="Close modal" />
<IconButton variant="close" aria-label="Dismiss" onClick={handleClose} />
```

#### IconButton Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `variant` | `'arrow' \| 'close'` | - | Icon type (required) |
| `aria-label` | `string` | - | Accessibility label (required) |
| `onClick` | `() => void` | - | Click handler |
| `disabled` | `boolean` | `false` | Disable button |

#### Hover Effects

| Variant | Effect | Duration |
|---------|--------|----------|
| `arrow` | Rotates -45° (points up-right) | 300ms |
| `close` | Rotates 180° | 600ms |

---

## 📁 Project Structure

```
├── app/                    # Next.js App Router pages
│   ├── globals.css         # Design system variables
│   ├── layout.tsx          # Root layout with fonts
│   └── page.tsx            # Home page
├── components/
│   └── ui/                 # UI components
│       ├── Button.tsx      # Text button component
│       ├── IconButton.tsx  # Icon button component
│       └── Card.tsx        # Card component
├── lib/                    # Utilities
├── types/                  # TypeScript types
└── README.md
```

---

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS](https://tailwindcss.com/docs)

---

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new).

Check out the [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
