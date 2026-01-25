'use client';

import { useState, useRef, useEffect } from 'react';
import Button from './Button';
import IconButton from './IconButtons';
import DatePicker from './DatePicker';
import TimePicker from './TimePicker';
import { Calendar, Clock, MapPin, ImagePlus } from 'lucide-react';
import Image from 'next/image';

interface CreateEventModalProps {
  isOpen: boolean;
  onClose: () => void;
  preview?: boolean;
}

export default function CreateEventModal({ isOpen, onClose, preview = false }: CreateEventModalProps) {
  const [title, setTitle] = useState('');
  const [date, setDate] = useState('');
  const [selectedDate, setSelectedDate] = useState<Date | undefined>();
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [time, setTime] = useState('');
  const [showTimePicker, setShowTimePicker] = useState(false);
  const [location, setLocation] = useState('');
  const [details, setDetails] = useState('');
  const [uploadedImage, setUploadedImage] = useState<{ file: File; preview: string } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const datePickerRef = useRef<HTMLDivElement>(null);
  const timePickerRef = useRef<HTMLDivElement>(null);

  // Close pickers when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (datePickerRef.current && !datePickerRef.current.contains(e.target as Node)) {
        setShowDatePicker(false);
      }
      if (timePickerRef.current && !timePickerRef.current.contains(e.target as Node)) {
        setShowTimePicker(false);
      }
    };
    if (showDatePicker || showTimePicker) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [showDatePicker, showTimePicker]);

  const handleDateSelect = (newDate: Date) => {
    setSelectedDate(newDate);
    const formatted = `${String(newDate.getDate()).padStart(2, '0')}/${String(newDate.getMonth() + 1).padStart(2, '0')}/${newDate.getFullYear()}`;
    setDate(formatted);
    setShowDatePicker(false);
  };

  const handleTimeSelect = (newTime: string) => {
    setTime(newTime);
    setShowTimePicker(false);
  };

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const preview = URL.createObjectURL(file);
      setUploadedImage({ file, preview });
    }
  };

  const handleRemoveImage = () => {
    if (uploadedImage) {
      URL.revokeObjectURL(uploadedImage.preview);
      setUploadedImage(null);
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  if (!isOpen && !preview) return null;

  const handleSubmit = () => {
    console.log({ title, date, time, location, details, uploadedImage: uploadedImage?.file });
    handleRemoveImage();
    onClose();
  };

  const modalContent = (
    <div className="relative w-full max-w-145 bg-white border border-parea-black shadow-[8px_8px_0_0_#000]">
      {/* Header */}
      <div className="relative h-17 border-b border-parea-black overflow-hidden bg-parea-white">
        <Image
          src="/modal-header-pattern.svg"
          alt=""
          fill
          className="object-cover"
        />
        <IconButton
          variant="close"
          onClick={onClose}
          aria-label="Close modal"
          className="absolute top-4 right-8 z-10"
        />
      </div>

      {/* Content */}
      <div className="p-8 min-h-146 flex flex-col">
        {/* Heading */}
        <h4 className="text-h4 mb-4">Create event</h4>

        {/* Form */}
        <div className="flex flex-col gap-6 flex-1">
          {/* Event Title */}
          <div className="flex flex-col gap-2">
            <label className="label">EVENT TITLE</label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Placeholder"
              className="w-full p-3 bg-parea-white border border-parea-black focus:outline-none"
            />
          </div>

          {/* Date, Time, Location Row */}
          <div className="flex gap-4">
            {/* Choose Date */}
            <div className="flex flex-col gap-2 flex-1" ref={datePickerRef}>
              <label className="label">CHOOSE DATE</label>
              <div className="relative">
                <input
                  type="text"
                  value={date}
                  onChange={(e) => setDate(e.target.value)}
                  onFocus={() => setShowDatePicker(true)}
                  placeholder="16/12/2025"
                  className="w-full p-3 pr-12 bg-parea-white border border-parea-border focus:outline-none cursor-pointer"
                  readOnly
                />
                <Calendar
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-parea-black pointer-events-none"
                  size={20}
                />
                {showDatePicker && (
                  <div className="absolute top-full left-0 mt-1 z-10">
                    <DatePicker
                      selectedDate={selectedDate}
                      onSelectDate={handleDateSelect}
                    />
                  </div>
                )}
              </div>
            </div>

            {/* Choose Hour */}
            <div className="flex flex-col gap-2 flex-1" ref={timePickerRef}>
              <label className="label">CHOOSE HOUR</label>
              <div className="relative">
                <input
                  type="text"
                  value={time}
                  onChange={(e) => setTime(e.target.value)}
                  onFocus={() => setShowTimePicker(true)}
                  placeholder="09:00 AM"
                  className="w-full p-3 pr-12 bg-parea-white border border-parea-border focus:outline-none cursor-pointer"
                  readOnly
                />
                <Clock
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-parea-black pointer-events-none"
                  size={20}
                />
                {showTimePicker && (
                  <div className="absolute top-full left-0 mt-1 z-10">
                    <TimePicker
                      selectedTime={time}
                      onSelectTime={handleTimeSelect}
                    />
                  </div>
                )}
              </div>
            </div>

            {/* Enter Location */}
            <div className="flex flex-col gap-2 flex-1">
              <label className="label">ENTER LOCATION</label>
              <div className="relative">
                <input
                  type="text"
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  placeholder="Enter Location"
                  className="w-full p-3 pr-12 bg-parea-white border border-parea-border focus:outline-none"
                />
                <MapPin
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-parea-black"
                  size={20}
                />
              </div>
            </div>
          </div>

          {/* Event Details */}
          <div className="flex flex-col gap-2 flex-1">
            <label className="label">EVENT DETAILS</label>
            <div className="relative">
              <textarea
                value={details}
                onChange={(e) => setDetails(e.target.value)}
                placeholder="What are you thinking?"
                className="w-full h-full min-h-35 p-4 bg-parea-white border border-parea-border focus:outline-none resize-none"
                style={{ fontFamily: 'var(--font-body), system-ui, sans-serif' }}
              />
              {/* Hidden file input */}
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleImageUpload}
                className="hidden"
              />
              {/* Image Upload Button */}
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="absolute bottom-4 right-3 hover:bg-parea-black/10 rounded transition-colors"
              >
                <ImagePlus size={24} stroke="#000" strokeWidth={1.5} />
              </button>
            </div>

            {/* Uploaded Image Preview */}
            {uploadedImage && (
              <div className="flex items-center gap-3 mt-3 p-3 border border-parea-border bg-parea-white">
                <Image
                  src={uploadedImage.preview}
                  alt="Uploaded preview"
                  width={40}
                  height={40}
                  className="object-cover"
                />
                <span className="flex-1 text-sm text-parea-black/60 truncate">
                  {uploadedImage.file.name}
                </span>
                <IconButton
                  variant="close"
                  size="sm"
                  transparent
                  onClick={handleRemoveImage}
                  aria-label="Remove image"
                />
              </div>
            )}
          </div>
        </div>

        {/* Submit Button */}
        <div className="flex justify-end mt-4">
          <Button variant="primary" size="lg" onClick={handleSubmit}>
            SUBMIT
          </Button>
        </div>
      </div>
    </div>
  );

  if (preview) {
    return modalContent;
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={onClose}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div onClick={(e) => e.stopPropagation()}>
        {modalContent}
      </div>
    </div>
  );
}
