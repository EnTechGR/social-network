'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import AuthLayout from '@/components/auth/AuthLayout';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import IconButton from '@/components/ui/IconButtons';
import { registerStep1, registerStep2 } from '@/lib/api';
import { isValidEmail, isValidPassword } from '@/lib/validations';
import Link from 'next/link';
import Image from 'next/image';

export default function SignUpPage() {
  const router = useRouter();
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

  const handleStep1Submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validation
    if (!firstName || !lastName || !dateOfBirth || !email || !password) {
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

    setIsLoading(true);

    try {
      const response = await registerStep1({
        email,
        password,
        first_name: firstName,
        last_name: lastName,
        date_of_birth: formatDateForAPI(dateOfBirth),
        gender: 'prefer_not_to_say', // Default, can be updated later
        avatar: avatar || undefined,
      });
      
      setRegisterData(response);
      setStep(2);
    } catch (err: any) {
      setError(err.message || 'Registration failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
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

          {/* Avatar Upload */}
          <div className="flex justify-center mb-8">
            <label className="cursor-pointer">
              <div className="w-32 h-32 rounded-lg border-4 border-parea-yellow overflow-hidden bg-linear-to-br from-parea-yellow to-parea-black relative">
                {avatarPreview ? (
                  <Image
                    src={avatarPreview}
                    alt="Avatar preview"
                    fill
                    className="object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <svg
                      width="48"
                      height="48"
                      viewBox="0 0 24 24"
                      fill="none"
                      className="text-parea-white opacity-50"
                    >
                      <path
                        d="M12 12C14.7614 12 17 9.76142 17 7C17 4.23858 14.7614 2 12 2C9.23858 2 7 4.23858 7 7C7 9.76142 9.23858 12 12 12Z"
                        stroke="currentColor"
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                      <path
                        d="M20.59 22C20.59 18.13 16.74 15 12 15C7.26 15 3.41 18.13 3.41 22"
                        stroke="currentColor"
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </div>
                )}
              </div>
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

            <Input
              type="text"
              placeholder={isDateFocused ? "MM/DD/YYYY" : "BIRTH DATE*"}
              value={dateOfBirth}
              onChange={handleDateChange}
              onFocus={() => setIsDateFocused(true)}
              onBlur={() => setIsDateFocused(false)}
              required
            />

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

          <div className="w-full py-4">
            <div className="w-full border-t border-parea-black h-[0.0625rem]"></div>
          </div>
          
          <Button
            type="button"
            variant="secondary"
            className="w-full flex items-center justify-center gap-2"
          >
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
              <path
                d="M17.64 9.20454C17.64 8.56636 17.5827 7.95272 17.4764 7.36363H9V10.845H13.8436C13.635 11.97 13.0009 12.9232 12.0477 13.5614V15.8195H15.9564C17.4382 14.4186 18.36 12.2727 18.36 9.20454Z"
                fill="#4285F4"
              />
              <path
                d="M9 18C11.43 18 13.467 17.1941 15.9564 15.8195L12.0477 13.5614C11.2418 14.1014 10.2109 14.4204 9 14.4204C6.65454 14.4204 4.67182 12.8373 3.96409 10.71H0.957275V13.0418C2.43818 15.9832 5.48182 18 9 18Z"
                fill="#34A853"
              />
              <path
                d="M3.96409 10.71C3.78409 10.17 3.68182 9.59318 3.68182 9C3.68182 8.40682 3.78409 7.29 3.96409 7.29V4.95818H0.957273C0.347727 6.17318 0 7.54773 0 9C0 10.4523 0.347727 11.8268 0.957273 13.0418L3.96409 10.71Z"
                fill="#FBBC05"
              />
              <path
                d="M9 3.57955C10.3214 3.57955 11.5077 4.03364 12.4405 4.92545L15.0218 2.34409C13.4632 0.891818 11.4259 0 9 0C5.48182 0 2.43818 2.01682 0.957275 4.95818L3.96409 7.29C4.67182 5.16273 6.65454 3.57955 9 3.57955Z"
                fill="#EA4335"
              />
            </svg>
            SIGN UP WITH GOOGLE
          </Button>

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
      // Complete registration step 2
      await registerStep2(registerData.id, {
        nickname,
        about_me: aboutMe,
      }, registerData.csrf_token);

      // Store auth data
      if (typeof window !== 'undefined') {
        localStorage.setItem('auth_token', registerData.csrf_token);
        localStorage.setItem('user_id', registerData.id);
      }

      router.push('/feed');
    } catch (err: any) {
      setError(err.message || 'Failed to complete registration. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthLayout>
      <div className="flex flex-col items-center justify-center h-screen px-8 py-4 md:py-8">
        <div className="w-full max-w-md">
          <h1 className="text-4xl font-bold text-parea-black mb-2 text-center">
            Sign Up
          </h1>
          <p className="text-regular text-parea-black mb-6 text-center">
            Find your people. Join the parea.
          </p>

          {/* Avatar Display */}
          {avatarPreview && (
            <div className="flex justify-center mb-8">
              <div className="w-32 h-32 rounded-lg border-4 border-parea-yellow overflow-hidden relative">
                <Image
                  src={avatarPreview}
                  alt="Avatar"
                  fill
                  className="object-cover"
                />
              </div>
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
                  min-h-[120px]
                  px-3 py-2
                  items-start
                  gap-2
                  self-stretch
                  rounded-[3rem]
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

          <div className="w-full py-4">
            <div className="w-full border-t border-parea-black h-[0.0625rem]"></div>
          </div>
          
          <Button
            type="button"
            variant="secondary"
            className="w-full flex items-center justify-center gap-2"
          >
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
              <path
                d="M17.64 9.20454C17.64 8.56636 17.5827 7.95272 17.4764 7.36363H9V10.845H13.8436C13.635 11.97 13.0009 12.9232 12.0477 13.5614V15.8195H15.9564C17.4382 14.4186 18.36 12.2727 18.36 9.20454Z"
                fill="#4285F4"
              />
              <path
                d="M9 18C11.43 18 13.467 17.1941 15.9564 15.8195L12.0477 13.5614C11.2418 14.1014 10.2109 14.4204 9 14.4204C6.65454 14.4204 4.67182 12.8373 3.96409 10.71H0.957275V13.0418C2.43818 15.9832 5.48182 18 9 18Z"
                fill="#34A853"
              />
              <path
                d="M3.96409 10.71C3.78409 10.17 3.68182 9.59318 3.68182 9C3.68182 8.40682 3.78409 7.29 3.96409 7.29V4.95818H0.957273C0.347727 6.17318 0 7.54773 0 9C0 10.4523 0.347727 11.8268 0.957273 13.0418L3.96409 10.71Z"
                fill="#FBBC05"
              />
              <path
                d="M9 3.57955C10.3214 3.57955 11.5077 4.03364 12.4405 4.92545L15.0218 2.34409C13.4632 0.891818 11.4259 0 9 0C5.48182 0 2.43818 2.01682 0.957275 4.95818L3.96409 7.29C4.67182 5.16273 6.65454 3.57955 9 3.57955Z"
                fill="#EA4335"
              />
            </svg>
            SIGN UP WITH GOOGLE
          </Button>

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
