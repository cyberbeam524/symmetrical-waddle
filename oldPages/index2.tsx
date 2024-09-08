import React, { useState, useEffect } from 'react';


const Home = () => {
    const [url, setUrl] = useState('');
    const [iframeUrl, setIframeUrl] = useState('');
    const [selectors, setSelectors] = useState([]);

    const handleUrlChange = e => setUrl(e.target.value);

    const handleLoadContent = () => {
      const encodedUrl = encodeURIComponent(url);
      setIframeUrl(`/api/proxy?url=${encodedUrl}`);
      // return `/api/proxy?url=${encodedUrl}`;
  };

  const addSelector = (selector) => {
      if (!selectors.includes(selector)) {
          setSelectors([...selectors, selector]);
      }
  };

  const removeSelector = (index) => {
      setSelectors(selectors.filter((_, idx) => idx !== index));
  };

    useEffect(() => {
      const handleMessage = event => {
          // Ensure security by checking origin or other properties
          if (typeof event.data === 'string') {
              addSelector(event.data);
          }
      };

      window.addEventListener('message', handleMessage);
      return () => window.removeEventListener('message', handleMessage);
  }, []);


    return (
        // <div>
        //     <h1>Web Content Embedder</h1>
        //     <input type="text" value={url} onChange={handleUrlChange} placeholder="Enter a URL" />
        //     <button onClick={handleLoadContent}>Load Content</button>
        //     <iframe src={iframeUrl} style={{ width: '100%', height: '80vh', border: '1px solid #ccc' }} />
        // </div>

        <div>
        <h1>Web Content Embedder</h1>
        <input type="text" value={url} onChange={handleUrlChange} placeholder="Enter a URL" />
        <button onClick={handleLoadContent}>Load Content</button>
        <iframe src={iframeUrl} style={{ width: '100%', height: '80vh' }} />

        <div>
            <h2>Selected CSS Selectors:</h2>
            <ul>
                {selectors.map((selector, index) => (
                    <li key={index}>
                        {selector} <button onClick={() => removeSelector(index)}>Remove</button>
                    </li>
                ))}
            </ul>
        </div>
    </div>
    );
};

export default Home;
