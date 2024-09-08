import React, { useState, useEffect } from 'react';

const Home = () => {
    const [url, setUrl] = useState('');
    const [iframeUrl, setIframeUrl] = useState('');
    const [selectors, setSelectors] = useState([]);

    const handleUrlChange = e => setUrl(e.target.value);

    const handleLoadContent = () => {
      const encodedUrl = encodeURIComponent(url);
      setIframeUrl(`/api/proxy?url=${encodedUrl}`);
    };

    const addSelector = (selector) => {
        if (!selectors.includes(selector)) {
            console.log(selectors, selector);
            setSelectors([...selectors, selector]);
        }
    };

    const removeSelector = (index) => {
        setSelectors(selectors.filter((_, idx) => idx !== index));
    };

    useEffect(() => {
        const handleMessage = event => {
            if (typeof event.data === 'string') {
                addSelector(event.data);
            }
        };

        window.addEventListener('message', handleMessage);
        return () => window.removeEventListener('message', handleMessage);
    }, []);

    return (
        <div className="container mx-auto p-4 bg-custom-white text-custom-black">
            <h1 className="text-3xl font-bold mb-4 bg-pattern-dot">Web Content Embedder</h1>
            <div className="mb-4">
                <input className="shadow appearance-none border rounded py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" 
                       type="text" value={url} onChange={handleUrlChange} placeholder="Enter a URL" />
                <button className="bg-custom-black hover:bg-gray-800 text-custom-white font-bold py-2 px-4 rounded ml-2"
                        onClick={handleLoadContent}>Load Content</button>
            </div>
            <iframe className="w-full h-80vh border-gray-300 border bg-pattern-stripe" src={iframeUrl}></iframe>

            <div className="mt-4">
                <h2 className="text-xl font-semibold">Selected CSS Selectors:</h2>
                <ul>
                    {selectors.map((selector, index) => (
                        <li key={index} className="flex justify-between items-center p-2 hover:bg-gray-100">
                            {selector} <button className="bg-custom-black hover:bg-gray-800 text-white font-bold py-1 px-3 rounded"
                                               onClick={() => removeSelector(index)}>Remove</button>
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    );
};

export default Home;

