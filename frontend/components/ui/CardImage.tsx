import React from 'react';

interface CardImageProps {
  type: 'post' | 'event';
  src?: string;
  alt?: string;
}

export const CardImage: React.FC<CardImageProps> = ({ 
  type, 
  src, 
  alt = 'Card image' 
}) => {
  // Determine which image to use based on type and whether src is provided
  const getImagePath = (): string => {
    if (src) {
      return src;
    }
    
    if (type === 'post') {
      return '/postDefaultImage.png';
    }
    
    return '/eventDefaultImage.png';
  };

  const imagePath = getImagePath();

  return (
    <div className="relative
      w-[300px]
      h-[274px]
      flex-shrink-0
      rounded-[4px]
      bg-black
      group/shadow">
      <div
        className="absolute
          w-[300px]
          h-[274px]
          rounded-[4px]
          bg-gray-300
          bg-center
          bg-cover
          bg-no-repeat
          transition-all
          duration-300
          top-0
          left-0
          group-hover/shadow:-left-2
          group-hover/shadow:-top-2"
        style={{
          backgroundImage: `url(${imagePath})`,
        }}
        role="img"
        aria-label={alt}
      />
    </div>
  );
};

export default CardImage;
