const Loading = () => {
    return (
      <div className="flex flex-col items-center justify-center flex-grow bg-gray-50">
        <div className="relative">
          {/* Outer spinning ring */}
          <div className="w-16 h-16 border-4 border-indigo-200 rounded-full animate-spin"></div>
  
          {/* Inner spinning ring */}
          <div className="absolute top-0 left-0 w-16 h-16 border-4 border-indigo-600 rounded-full animate-spin border-t-transparent"></div>
  
          {/* Inner pulsing circle */}
          <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
            <div className="w-8 h-8 bg-indigo-100 rounded-full animate-pulse"></div>
          </div>
        </div>
  
        {/* Loading text */}
        <div className="mt-4 text-indigo-600 font-medium animate-pulse">
          Cargando...
        </div>
      </div>
    );
  };
export default Loading;
