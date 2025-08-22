  import Link from 'next/link';
  import { Home } from 'lucide-react';
  import { useState } from 'react';

  const HomeIcon = () => {
    const [isHovered, setIsHovered] = useState(false);

    return (
      <Link 
        href="/" 
        className="relative inline-flex items-center justify-center w-12 h-12
                  group perspective-1000"
        aria-label="Ir al inicio"
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        {/* Efecto de fondo */}
        <div className={`
          absolute inset-0 rounded-full bg-gradient-to-tr from-blue-500 to-purple-500
          transform transition-all duration-300 ease-out
          ${isHovered ? 'opacity-100 scale-105' : 'opacity-0 scale-95'}
        `} />
        
        {/* Círculo interior */}
        <div className={`
          absolute inset-0.5 rounded-full bg-white
          transform transition-all duration-300 ease-out
          ${isHovered ? 'scale-100' : 'scale-95'}
        `} />
        
        {/* Icono */}
        <Home 
          className={`
            relative w-6 h-6 transform transition-all duration-300 ease-out hover:scale-110 text-blue-600
          `}
          strokeWidth={isHovered ? 2.5 : 2}
        />

        {/* Tooltip */}
        <div className={`
          absolute -bottom-8 left-1/2 transform -translate-x-1/2
          px-2 py-1 bg-gray-900 text-white text-xs rounded-md whitespace-nowrap
          transition-all duration-200
          ${isHovered ? 'opacity-100 translate-y-0' : 'opacity-0 -translate-y-1'}
        `}>
          Inicio
        </div>
      </Link>
    );
  };

  export default HomeIcon;