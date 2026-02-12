import EventDetailFrame from '@/components/ui/EventDetailFrame';

export default function EventPage({ params }: { params: Promise<{ id: string }> }) {
  return (
    <div className="min-h-screen bg-parea-white px-16 py-12">
      <div className="w-full">
        <EventDetailFrame
          title="24-Hour Minimal Techno"
          imageSrc="/test-event.png"
          location="LOCATION"
          goingCount={24}
          date="11 JAN 2022"
          time="09:00 AM"
          eventText="24-Hour Minimal Techno is an immersive experience that brings together DJs and electronic music lovers for a full day and night of minimal techno. The event features a carefully curated lineup of local and international artists, creating a journey through deep, hypnotic beats and atmospheric soundscapes. From the opening set at noon to the closing sunrise session, the energy builds and evolves, offering something for both longtime fans and newcomers to the genre."
        />
      </div>
    </div>
  );
}
