import React from 'react';
import Image from 'next/image';

interface CardImageProps {
  type: 'post';
  src?: string;
  alt?: string;
  /** Set true for the first card in a feed (above the fold) to fix LCP */
  priority?: boolean;
}

export const CardImage: React.FC<CardImageProps> = ({
  type,
  src,
  alt = 'Card image',
  priority = false,
}) => {
  const getImagePath = (): string => {
    if (src) return src;
    if (type === 'post') return '/postDefaultImage.png';
    return '/postDefaultImage.png';
  };

  const imagePath = getImagePath();
  const isExternal = imagePath.startsWith('http://') || imagePath.startsWith('https://');

  const containerClass =
    'relative w-[300px] h-[274px] flex-shrink-0 rounded-[4px] bg-black group/shadow';
  const innerClass =
    'absolute w-[300px] h-[274px] rounded-[4px] bg-gray-300 bg-center bg-cover bg-no-repeat transition-all duration-300 top-0 left-0 group-hover/shadow:-left-2 group-hover/shadow:-top-2';

  if (isExternal) {
    return (
      <div className={containerClass}>
        <Image
          src={imagePath}
          alt={alt}
          fill
          className="object-cover rounded-[4px] object-center"
          sizes="300px"
          priority={priority}
        />
      </div>
    );
  }

  return (
    <div className={containerClass}>
      <div
        className={innerClass}
        style={{ backgroundImage: `url(${imagePath})` }}
        role="img"
        aria-label={alt}
      />
    </div>
  );
};

export default CardImage;
