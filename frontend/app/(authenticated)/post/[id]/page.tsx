import PostDetailFrame from '@/components/ui/PostDetailFrame';
import type { CommentItem } from '@/components/ui/CommentHolder';

const sampleComments: CommentItem[] = [
  {
    id: '1',
    avatarSrc: '/user-avatar-default.png',
    avatarAlt: 'Bessie Cooper',
    userName: 'BESSIE COOPER',
    userDate: '11 Jan 2022',
    text: 'Woaw! Such a cool collection. Thank you Olii!!!',
    likeCount: 2,
  },
  {
    id: '2',
    avatarSrc: '/user-avatar-default.png',
    avatarAlt: 'Jacob Jones',
    userName: 'JACOB JONES',
    userDate: '11 Jan 2022',
    text: 'Love these hidden spots. Need to visit soon.',
    likeCount: 1,
  },
];

export default function PostPage({ params }: { params: Promise<{ id: string }> }) {
  return (
    <div className="min-h-screen bg-parea-white px-16 py-12">
      <div className="w-full">
        <PostDetailFrame
          title="Local Explorer Discovers Hidden Urban Gems"
          avatarSrc="/test-avatar.png"
          avatarAlt="Olivia Minter"
          userName="OLIVIA MINTER"
          userDate="02 Jan 2025"
          likeCount={250}
          commentCount={4}
          imageSrc="/test-post.png"
          postText="Wandering through the city streets, a local explorer stumbled upon forgotten alleyways and hidden courtyards that most residents never see. What started as a casual afternoon walk turned into an unexpected journey through the city's secret corners—each turn revealing another piece of urban history waiting to be discovered."
          comments={sampleComments}
        />
      </div>
    </div>
  );
}
