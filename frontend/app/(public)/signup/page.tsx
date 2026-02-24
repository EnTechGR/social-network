'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import AuthLayout from '@/components/auth/AuthLayout';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import IconButton from '@/components/ui/IconButtons';
import { registerStep1 } from '@/lib/api';
import { isValidEmail, isValidPassword } from '@/lib/validations';
import Link from 'next/link';
import Avatar from '@/components/ui/Avatar';

export default function SignUpPage() {
  const [step, setStep] = useState(1);
  const [avatar, setAvatar] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [dateOfBirth, setDateOfBirth] = useState('');
  const [isDateFocused, setIsDateFocused] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [registerData, setRegisterData] = useState<any>(null);
  const [gender, setGender] = useState<'male' | 'female' | 'other' | 'prefer_not_to_say' | ''>('');

  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setAvatar(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setAvatarPreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleDateChange = (value: string) => {
    // Remove all non-digit characters
    const digits = value.replace(/\D/g, '');

    // Limit to 8 digits (MMDDYYYY)
    const limited = digits.slice(0, 8);

    // Format with slashes: MM/DD/YYYY
    let formatted = '';
    if (limited.length > 0) {
      formatted = limited.slice(0, 2);
      if (limited.length > 2) {
        formatted += '/' + limited.slice(2, 4);
      }
      if (limited.length > 4) {
        formatted += '/' + limited.slice(4, 8);
      }
    }

    setDateOfBirth(formatted);
  };

  // Convert MM/DD/YYYY to YYYY-MM-DD for API
  const formatDateForAPI = (dateStr: string): string => {
    if (!dateStr || dateStr.length !== 10) return dateStr;
    const [month, day, year] = dateStr.split('/');
    if (month && day && year && year.length === 4) {
      return `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
    }
    return dateStr;
  };

  const handleStep1Submit = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validation
    if (!firstName || !lastName || !dateOfBirth || !gender || !email || !password) {
      setError('Please fill in all required fields');
      return;
    }

    if (!isValidEmail(email)) {
      setError('Please enter a valid email address');
      return;
    }

    if (!isValidPassword(password)) {
      setError('Password must be at least 8 characters with uppercase, lowercase, and a number');
      return;
    }

    // Save the data locally and proceed to step 2
    setRegisterData({
      email,
      password,
      first_name: firstName,
      last_name: lastName,
      date_of_birth: formatDateForAPI(dateOfBirth),
      gender,
      avatar: avatar || undefined,
    });

    setStep(2);
  };

  if (step === 2) {
    return (
      <SignUpStep2
        registerData={registerData}
        avatarPreview={avatarPreview}
        onBack={() => setStep(1)}
      />
    );
  }

  return (
    <AuthLayout>
      <div className="flex flex-col items-center justify-center w-full px-8 py-4">
        <div className="w-full max-w-md">
          <div className="flex flex-col items-center gap-4 mb-8">
            <h2 className="text-4xl font-bold text-parea-black">
              Sign Up
            </h2>
            <p className="text-regular text-parea-black">
              Find your people. Join the parea.
            </p>
          </div>

          {/* Avatar Upload - always use Avatar (dots, yellow bg, frame); empty state shows default user image */}
          <div className="flex justify-center mb-8">
            <label className="cursor-pointer block">
              <Avatar
                type="image"
                src={avatarPreview ?? undefined}
                alt="Avatar preview"
                size="profile"
                isSelf
              />
              <input
                type="file"
                accept="image/*"
                onChange={handleAvatarChange}
                className="hidden"
              />
            </label>
          </div>

          <form onSubmit={handleStep1Submit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <Input
                type="text"
                placeholder="FIRST NAME*"
                value={firstName}
                onChange={setFirstName}
                required
              />
              <Input
                type="text"
                placeholder="LAST NAME*"
                value={lastName}
                onChange={setLastName}
                required
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Input
                type="text"
                placeholder={isDateFocused ? "MM/DD/YYYY" : "BIRTH DATE*"}
                value={dateOfBirth}
                onChange={handleDateChange}
                onFocus={() => setIsDateFocused(true)}
                onBlur={() => setIsDateFocused(false)}
                required
              />
              <select
                id="gender"
                value={gender}
                onChange={(e) => setGender(e.target.value as 'male' | 'female' | 'other' | 'prefer_not_to_say' | '')}
                className={`select-chevron w-full h-input rounded-button border border-parea-black px-4 py-2 bg-parea-white text-regular focus:outline-none cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed uppercase font-medium ${
                  gender === '' ? 'text-[#00000099]' : 'text-parea-black'
                }`}
                style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                required
              >
                <option value="" disabled>GENDER*</option>
                <option value="prefer_not_to_say">Prefer not to say</option>
                <option value="male">Male</option>
                <option value="female">Female</option>
                <option value="other">Other</option>
              </select>
            </div>

            <Input
              type="email"
              placeholder="EMAIL*"
              value={email}
              onChange={setEmail}
              required
            />

            <Input
              type="password"
              placeholder="PASSWORD*"
              value={password}
              onChange={setPassword}
              required
            />

            {error && (
              <div className="text-sm text-red-600 text-center">{error}</div>
            )}

            <div className="flex justify-end">
              <IconButton
                variant="arrow-right"
                text="NEXT"
                onClick={(e) => {
                  if (e) {
                    e.preventDefault();
                    handleStep1Submit(e as any);
                  }
                }}
                disabled={isLoading}
                aria-label="Next"
              />
            </div>
          </form>

          <div className="mt-6 text-center">
            <p className="text-regular text-parea-black">
              Already have an account?{' '}
              <Link href="/login" className="underline">
                Log In
              </Link>
            </p>
          </div>
        </div>
      </div>
    </AuthLayout>
  );
}

// Step 2 Component
function SignUpStep2({
  registerData,
  avatarPreview,
  onBack,
}: {
  registerData: any;
  avatarPreview: string | null;
  onBack: () => void;
}) {
  const router = useRouter();
  const [nickname, setNickname] = useState('');
  const [aboutMe, setAboutMe] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!nickname) {
      setError('Please enter a username');
      return;
    }

    setIsLoading(true);

    try {
      // Send all registration data in a single request
      const response = await registerStep1({
        ...registerData,
        nickname,
        about_me: aboutMe,
      });

      // Store CSRF token and user data (actual auth is via HttpOnly cookie)
      if (typeof window !== 'undefined') {
        localStorage.setItem('csrf_token', response.csrf_token);
        localStorage.setItem('user_data', JSON.stringify(response.user));
      }

      router.push('/feed');
    } catch (err: any) {
      setError(err.message || 'Failed to create account. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthLayout>
      <div className="flex flex-col items-center justify-center w-full px-8 py-4">
        <div className="w-full max-w-md">
          <div className="flex flex-col items-center gap-4 mb-8">
            <h1 className="text-4xl font-bold text-parea-black text-center">
              Sign Up
            </h1>
            <p className="text-regular text-parea-black text-center">
              Find your people. Join the parea.
            </p>
          </div>

          {/* Avatar Display */}
          {avatarPreview && (
            <div className="flex justify-center mb-8">
              <Avatar
                type="image"
                src={avatarPreview}
                alt="Avatar"
                size="profile"
                isSelf
              />
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              type="text"
              placeholder="USERNAME"
              value={nickname}
              onChange={setNickname}
              required
            />

            <div className="flex flex-col gap-1.5">
              <textarea
                placeholder="BIO"
                value={aboutMe}
                onChange={(e) => setAboutMe(e.target.value)}
                className="
                  flex
                  min-h-30
                  px-4 py-2
                  items-start
                  gap-2
                  self-stretch
                  rounded-input
                  border
                  border-parea-black
                  bg-parea-white
                  text-parea-black
                  placeholder:text-[#00000099]
                  placeholder:font-mono
                  placeholder:text-regular
                  placeholder:leading-relaxed
                  placeholder:uppercase
                  focus:outline-none
                  resize-none
                "
                style={{
                  fontFamily: 'var(--font-body), system-ui, sans-serif',
                }}
              />
            </div>

            {error && (
              <div className="text-sm text-red-600 text-center">{error}</div>
            )}

            <div className="flex items-center justify-between gap-4 mt-6">
              <IconButton
                variant="arrow-left"
                text="PREVIOUS"
                onClick={onBack}
                aria-label="Previous"
              />

              <IconButton
                variant="arrow-right"
                text="CREATE ACCOUNT"
                onClick={() => {
                  const form = document.querySelector('form');
                  if (form) {
                    handleSubmit(new Event('submit') as any);
                  }
                }}
                disabled={isLoading}
                aria-label="Create account"
              />
            </div>
          </form>

          <div className="mt-6 text-center">
            <p className="text-regular text-parea-black">
              Already have an account?{' '}
              <Link href="/login" className="underline">
                Log In
              </Link>
            </p>
          </div>
        </div>
      </div>
    </AuthLayout>
  );
}
